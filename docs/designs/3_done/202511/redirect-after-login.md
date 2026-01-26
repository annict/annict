# ログイン後リダイレクト機能 設計書

## 概要

ログインが必要なページにアクセスした際、ログインページへリダイレクトし、ログイン完了後に元のページへ自動的に戻る機能を実装する。

**目的**:

- ユーザーがログイン後に元々アクセスしようとしていたページにスムーズに戻れるようにする
- OAuthの認可フロー（`/oauth/authorize`）などでログインが必要な場合に、認可フローを中断せずに継続できるようにする

**背景**:

- 現状、ログインが必要なページからログインページに遷移した場合、ログイン後は常にトップページ（`/`）にリダイレクトされる
- OAuthクライアントの認可フローでは、ログイン後に元の認可URLに戻る必要がある
- Rails版では `back` パラメータを使用しているため、Go版も同じパラメータ名を使用する

## 要件

### 機能要件

- ログインページに `back` パラメータ（URLエンコードされたパス）を付けてアクセスできる
  - 例: `/sign_in?back=%2Foauth%2Fauthorize%3Fclient_id%3DWpVu_xxxxU`
- ログイン成功後、`back` パラメータで指定されたパスにリダイレクトする
- `back` パラメータがない場合、または無効な場合は、デフォルトのリダイレクト先（`/`）にリダイレクトする
- 認証ミドルウェアで未認証ユーザーをログインページにリダイレクトする際、元のURLを `back` パラメータとして付与する

### 非機能要件

#### セキュリティ（オープンリダイレクト脆弱性対策）

`back` パラメータを使用したリダイレクトは、オープンリダイレクト脆弱性のリスクがある。以下の対策を実施する：

- **パスのみを許可**: `back` パラメータは `/` で始まる相対パスのみを許可
- **絶対URLを拒否**: `http://`、`https://` で始まるURLは拒否
- **プロトコル相対URLを拒否**: `//example.com` 形式のURLは拒否
- **バリデーション失敗時**: 無効な `back` パラメータの場合はデフォルトのリダイレクト先（`/`）を使用

## 設計

### 2つのログインフロー

現在のログインには2つのフローがあり、両方で `back` パラメータを引き継ぐ必要がある：

**フロー1: メール認証コードログイン（パスワードなしユーザー）**
```
GET /sign_in → POST /sign_in → GET /sign_in/code → POST /sign_in/code
```

**フロー2: パスワードログイン（パスワードありユーザー）**
```
GET /sign_in → POST /sign_in → GET /sign_in/password → POST /sign_in/password
```

### 処理フロー（メール認証コードログイン）

```
1. ユーザーが認証が必要なページ（例: /oauth/authorize?client_id=xxx）にアクセス
2. 認証ミドルウェアが未認証を検出
3. /sign_in?back=%2Foauth%2Fauthorize%3Fclient_id%3Dxxx にリダイレクト

[ステップ1: GET /sign_in?back=...]
4. ログインページで back パラメータを hidden フィールドに保持
5. ユーザーがメールアドレスを入力して送信

[ステップ2: POST /sign_in]
6. ユーザーにパスワードがない場合、認証コードをメール送信
7. /sign_in/code?back=... にリダイレクト（back パラメータを引き継ぐ）

[ステップ3: GET /sign_in/code?back=...]
8. 認証コード入力ページで back パラメータを hidden フィールドに保持
9. ユーザーが認証コードを入力して送信

[ステップ4: POST /sign_in/code]
10. 認証コードを検証してログイン完了
11. back パラメータの値（/oauth/authorize?client_id=xxx）にリダイレクト
```

### 処理フロー（パスワードログイン）

```
[ステップ2: POST /sign_in] の分岐
6. ユーザーにパスワードがある場合
7. /sign_in/password?back=... にリダイレクト（back パラメータを引き継ぐ）

[ステップ3: GET /sign_in/password?back=...]
8. パスワード入力ページで back パラメータを hidden フィールドに保持
9. ユーザーがパスワードを入力して送信

[ステップ4: POST /sign_in/password]
10. パスワードを検証してログイン完了
11. back パラメータの値（/oauth/authorize?client_id=xxx）にリダイレクト
```

### back パラメータの引き継ぎ

| ステップ | 入力 | 出力 |
|----------|------|------|
| GET /sign_in | URLパラメータ `?back=...` | hidden フィールド `<input name="back">` |
| POST /sign_in | フォーム `back` | リダイレクト先 `/sign_in/code?back=...` または `/sign_in/password?back=...` |
| GET /sign_in/code | URLパラメータ `?back=...` | hidden フィールド `<input name="back">` |
| POST /sign_in/code | フォーム `back` | リダイレクト先 `back` の値 |
| GET /sign_in/password | URLパラメータ `?back=...` | hidden フィールド `<input name="back">` |
| POST /sign_in/password | フォーム `back` | リダイレクト先 `back` の値 |

### コード設計

#### 1. リダイレクトURLバリデーション関数

`internal/redirect/` パッケージを新規作成し、リダイレクトURL のバリデーションロジックを実装する。

```go
// internal/redirect/redirect.go
package redirect

// ValidateBackURL は back パラメータの値が安全かどうかを検証する
// 安全な場合は true を返し、危険な場合は false を返す
func ValidateBackURL(backURL string) bool {
    // 空文字の場合は無効
    if backURL == "" {
        return false
    }

    // "/" で始まらない場合は無効（相対パスのみ許可）
    if !strings.HasPrefix(backURL, "/") {
        return false
    }

    // "//" で始まる場合は無効（プロトコル相対URL）
    if strings.HasPrefix(backURL, "//") {
        return false
    }

    return true
}

// GetSafeRedirectURL は安全なリダイレクトURLを返す
// backURL が無効な場合はデフォルトURL（"/"）を返す
func GetSafeRedirectURL(backURL string) string {
    if ValidateBackURL(backURL) {
        return backURL
    }
    return "/"
}
```

#### 2. 認証ミドルウェアの修正

`internal/middleware/auth.go` を修正し、未認証ユーザーをリダイレクトする際に `back` パラメータを付与する。

```go
// RequireAuth は認証を必要とするミドルウェア
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ... 認証チェック ...

        if !authenticated {
            // 現在のURLを back パラメータとして付与
            backURL := r.URL.RequestURI()
            redirectURL := "/sign_in?back=" + url.QueryEscape(backURL)
            http.Redirect(w, r, redirectURL, http.StatusFound)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

#### 3. ログインハンドラーの修正

以下のハンドラーを修正する：

**sign_in/new.go（ログインページ表示）**:
- `back` パラメータを取得してテンプレートに渡す

**sign_in/create.go（認証方式の分岐）**:
- フォームから `back` パラメータを取得
- `/sign_in/code?back=...` または `/sign_in/password?back=...` にリダイレクトする際に `back` パラメータを引き継ぐ

**sign_in_code/show.go（認証コード入力ページ表示）**:
- `back` パラメータを取得してテンプレートに渡す

**sign_in_code/create.go（認証コード検証＆ログイン）**:
- フォームから `back` パラメータを取得
- `redirect.GetSafeRedirectURL()` でバリデーション
- ログイン成功後、安全なリダイレクト先にリダイレクト
- **既存の `redirect_to` パラメータを `back` に変更**

**sign_in_password/new.go（パスワード入力ページ表示）**:
- `back` パラメータを取得してテンプレートに渡す

**sign_in_password/create.go（パスワード検証＆ログイン）**:
- フォームから `back` パラメータを取得
- `redirect.GetSafeRedirectURL()` でバリデーション
- ログイン成功後、安全なリダイレクト先にリダイレクト
- **既存の `redirect_to` パラメータを `back` に変更**

#### 4. テンプレートの修正

以下のテンプレートを修正し、`back` パラメータを hidden フィールドに保持する。

**sign_in/new.templ**:
```html
<input type="hidden" name="back" value={ backURL } />
```

**sign_in_code/show.templ**:
```html
<input type="hidden" name="back" value={ backURL } />
```

**sign_in_password/show.templ**:
```html
<input type="hidden" name="back" value={ backURL } />
```

### 既存コードとの互換性

既存の `sign_in_code/create.go` と `sign_in_password/create.go` では `redirect_to` パラメータを使用しているが、Rails版との互換性のため `back` パラメータに変更する。

変更箇所：
```go
// Before
redirectTo := r.FormValue("redirect_to")

// After
backURL := r.FormValue("back")
redirectTo := redirect.GetSafeRedirectURL(backURL)
```

### ファイル構成

```
internal/
├── redirect/
│   ├── redirect.go       # 新規: リダイレクトURLバリデーション
│   └── redirect_test.go  # 新規: テスト
├── middleware/
│   └── auth.go           # 修正: back パラメータ付与
├── handler/
│   ├── sign_in/
│   │   ├── new.go        # 修正: back パラメータをテンプレートに渡す
│   │   └── create.go     # 修正: リダイレクト時に back を引き継ぐ（2箇所）
│   ├── sign_in_code/
│   │   ├── show.go       # 修正: back パラメータをテンプレートに渡す
│   │   └── create.go     # 修正: redirect_to → back に変更、バリデーション追加
│   └── sign_in_password/
│       ├── new.go        # 修正: back パラメータをテンプレートに渡す
│       └── create.go     # 修正: redirect_to → back に変更、バリデーション追加
└── templates/
    └── pages/
        ├── sign_in/
        │   └── new.templ           # 修正: hidden フィールド追加
        ├── sign_in_code/
        │   └── show.templ          # 修正: hidden フィールド追加
        └── sign_in_password/
            └── show.templ          # 修正: hidden フィールド追加
```

## タスクリスト

### フェーズ 1: リダイレクトURLバリデーション

- [x] **1-1**: リダイレクトURLバリデーション関数の実装

  - `internal/redirect/redirect.go` を新規作成
  - `ValidateBackURL()` 関数の実装
  - `GetSafeRedirectURL()` 関数の実装
  - セキュリティテスト（オープンリダイレクト攻撃のテストケース）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 50 行 + テスト 100 行）

### フェーズ 2: ログインページの修正（sign_in）

- [x] **2-1**: sign_in ハンドラーとテンプレートの修正

  - `internal/handler/sign_in/new.go` の修正（back パラメータをテンプレートに渡す）
  - `internal/handler/sign_in/create.go` の修正（2箇所のリダイレクトで back を引き継ぐ: /sign_in/code と /sign_in/password）
  - `internal/templates/pages/sign_in/new.templ` の修正（hidden フィールド追加）
  - 統合テストの追加
  - **想定ファイル数**: 約 5 ファイル（実装 3 + テスト 2）
  - **想定行数**: 約 150 行（実装 50 行 + テスト 100 行）

### フェーズ 3: 認証コード入力ページの修正（sign_in_code）

- [x] **3-1**: sign_in_code ハンドラーとテンプレートの修正

  - `internal/handler/sign_in_code/show.go` の修正（back パラメータをテンプレートに渡す）
  - `internal/handler/sign_in_code/create.go` の修正（redirect_to → back に変更、バリデーション追加）
  - `internal/templates/pages/sign_in_code/show.templ` の修正（hidden フィールド追加）
  - 統合テストの追加
  - **想定ファイル数**: 約 5 ファイル（実装 3 + テスト 2）
  - **想定行数**: 約 150 行（実装 50 行 + テスト 100 行）

### フェーズ 4: パスワード入力ページの修正（sign_in_password）

- [x] **4-1**: sign_in_password ハンドラーとテンプレートの修正

  - `internal/handler/sign_in_password/new.go` の修正（back パラメータをテンプレートに渡す）
  - `internal/handler/sign_in_password/create.go` の修正（redirect_to → back に変更、バリデーション追加）
  - `internal/templates/pages/sign_in_password/show.templ` の修正（hidden フィールド追加）
  - 統合テストの追加
  - **想定ファイル数**: 約 5 ファイル（実装 3 + テスト 2）
  - **想定行数**: 約 150 行（実装 50 行 + テスト 100 行）

### フェーズ 5: 認証ミドルウェアの修正

- [x] **5-1**: 認証ミドルウェアで back パラメータを付与

  - `internal/middleware/auth.go` の修正（RequireAuth で back パラメータを付与）
  - 統合テストの追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 100 行（実装 30 行 + テスト 70 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **セッションへの保存**: `back` パラメータをセッションに保存する方式（シンプルさを優先し、URLパラメータとフォームのhiddenフィールドで完結させる）
- **外部URLへのリダイレクト**: 外部ドメインへのリダイレクトは許可しない（セキュリティリスク）
- **サインアップページでの対応**: サインアップページでの同様の機能（必要に応じて別タスクで対応）

## 参考資料

- [OWASP - Unvalidated Redirects and Forwards](https://cheatsheetseries.owasp.org/cheatsheets/Unvalidated_Redirects_and_Forwards_Cheat_Sheet.html)
- [Rails の redirect_back_or_default パターン](https://api.rubyonrails.org/classes/ActionController/Redirecting.html)
