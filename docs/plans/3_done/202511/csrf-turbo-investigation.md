# CSRF トークンと Turbo Drive の互換性問題調査

## 問題の概要

### 症状

- Rails 版のページから Rails 版のページに Turbo Drive で遷移するときに `ArgumentError - inputs must be of equal length` エラーが発生
- Turbo を `Turbo.session.drive = false;` で無効にするとエラーは発生しない
- ページリロードでもエラーは発生しない
- Rails 版のページは Go 版のリバースプロキシを介して表示している

### エラーの発生箇所

`ArgumentError - inputs must be of equal length` は Rails の `ActionController::RequestForgeryProtection` モジュール内の `xor_byte_strings` メソッドで発生する。このメソッドは CSRF トークンのマスク/アンマスク処理で使用される。

```ruby
# Rails の実装
def xor_byte_strings(s1, s2)
  if s1.bytesize != s2.bytesize
    raise ArgumentError, "inputs must be of equal length"
  end
  # ...
end
```

## 調査済み項目

### 1. Go 版の CSRF トークン生成

**ファイル**: `go/internal/session/csrf.go`

```go
func GenerateCSRFToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(b), nil
}
```

- 32 バイトのランダムデータを生成
- `base64.StdEncoding` で Base64 エンコード（標準 Base64、`+` と `/` を使用）
- 結果は約 44 文字の文字列

### 2. Go 版の CSRF ミドルウェア

**ファイル**: `go/internal/middleware/csrf.go`

- GET/HEAD/OPTIONS リクエストはチェックをスキップ
- フォームの `csrf_token` パラメータまたは `X-CSRF-Token` ヘッダーからトークンを取得
- セッションの CSRFトークンと**単純な文字列比較**で検証
- **マスクトークンの検証はサポートしていない**

### 3. Rails 版の CSRF 設定

**ファイル**: `rails/config/initializers/per_form_csrf_tokens.rb`

```ruby
Rails.application.config.action_controller.per_form_csrf_tokens = true
```

- **per_form_csrf_tokens が有効**
- 各フォームごとに異なる CSRF トークンを生成
- トークンはフォームの action と method に基づいて HMAC で生成される

### 4. Rails 版のセッションストア

**ファイル**: `rails/config/initializers/session_store.rb`

```ruby
Annict::Application.config.session_store :active_record_store,
  key: "_annict_session_v201904",
  domain: ENV["ANNICT_COOKIE_DOMAIN"].presence,
  expire_after: 30.days
```

**ファイル**: `rails/config/application.rb`

```ruby
ActiveRecord::SessionStore::Session.serializer = :null
```

- `active_record_store` を使用（PostgreSQL の sessions テーブル）
- `serializer = :null` を使用（データをそのまま保存）
- `data` カラムは `jsonb` 型

### 5. Go 版のリバースプロキシ

**ファイル**: `go/internal/middleware/reverse_proxy.go`

- `httputil.ReverseProxy` を使用
- Go 版で処理するパスはホワイトリストで定義
- それ以外のパスは Rails 版にプロキシ
- **レスポンスは変更していない**（`ModifyResponse` はログ出力のみ）

### 6. ミドルウェアの適用順序

**ファイル**: `go/cmd/server/main.go`

```go
// ミドルウェアの適用順序
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)
r.Use(middleware.RequestID)
r.Use(middleware.RealIP)

// リバースプロキシミドルウェア（最優先で配置）
if reverseProxyMW != nil {
    r.Use(reverseProxyMW.Middleware)
}

// 以下はGo版で処理する場合のみ適用される
r.Use(authMiddleware.MethodOverride)
r.Use(authMW.Middleware)
r.Use(i18n.Middleware)
r.Use(csrfMiddleware.Middleware)
```

- リバースプロキシミドルウェアが最初に実行される
- Rails 版にプロキシする場合、`next` を呼ばないため後続のミドルウェアはスキップされる

### 7. セッションテーブルの構造

**ファイル**: `rails/db/structure.sql`

```sql
CREATE TABLE public.sessions (
    id bigint NOT NULL,
    session_id character varying NOT NULL,
    data jsonb DEFAULT '"{}"'::jsonb NOT NULL,
    ...
);
```

- `data` カラムは `jsonb` 型
- Go 版も Rails 版も同じテーブルを使用

### 8. Go 版と Rails 版の CSRF トークン形式の違い

| 項目                   | Go 版                      | Rails 版                     |
| ---------------------- | -------------------------- | ---------------------------- |
| セッション保存         | Base64 文字列（44文字）    | Base64 文字列（44文字）      |
| HTML 出力（meta タグ） | なし                       | マスクトークン（88文字程度） |
| フォーム埋め込み       | 非マスクトークン（44文字） | マスクトークン（88文字程度） |
| 検証方法               | 単純な文字列比較           | マスク解除後に比較           |

### 9. Rails の CSRF トークンマスク処理

```ruby
# マスク処理
def mask_token(raw_token)
  one_time_pad = SecureRandom.random_bytes(32)
  encrypted_csrf_token = xor_byte_strings(one_time_pad, raw_token)
  Base64.strict_encode64(one_time_pad + encrypted_csrf_token)
end

# アンマスク処理
def unmask_token(masked_token)
  one_time_pad = masked_token[0...32]
  encrypted_csrf_token = masked_token[32..-1]
  xor_byte_strings(one_time_pad, encrypted_csrf_token)
end
```

- `mask_token` は 32 バイトの raw_token を期待
- `unmask_token` は 64 バイトの masked_token を期待
- 入力が期待されるサイズでない場合、`xor_byte_strings` でエラーが発生

## 根本原因（解決済み）

### 問題の発見

セッションデータを確認したところ、`_csrf_token` が**空文字列**であることが判明：

```json
{
  "_csrf_token": "",
  "warden.user.user.key": [[320394], "$2a$11$xxxxxxxxxxxxxxxxxxxxx"]
}
```

### 原因

**ファイル**: `go/internal/usecase/create_session.go`（60行目）

```go
// セッションデータを作成（Railsのwarden形式）
sessionData := map[string]any{
    "warden.user.user.key": []any{
        []any{userID},
        authenticatableSalt,
    },
    "_csrf_token": "", // ← ここが問題！空文字列を設定していた
}
```

Go 版のログインページでセッションを作成する際に、CSRF トークンを空文字列として保存していた。
コメントには「CSRFトークンは後で設定可能」とあったが、実際には設定されないまま使用されていた。

### エラーが発生する流れ

1. ユーザーが Go 版のログインページ（`/sign_in`）でログイン
2. Go 版が `CreateSessionUsecase.Execute()` でセッションを作成
3. `_csrf_token` が空文字列としてDBに保存される
4. ログイン後、Rails 版のページにリダイレクト
5. Rails が `csrf_meta_tags` をレンダリング
6. `real_csrf_token(session)` で `_csrf_token`（空文字列）を Base64 デコード
7. デコード結果が 0 バイト（期待値は 32 バイト）
8. `mask_token` で XOR 演算時に長さ不一致エラー

### 修正内容

`go/internal/usecase/create_session.go` で、空文字列ではなく実際の CSRF トークンを生成するように修正：

```go
// Rails互換のCSRFトークンを生成
csrfToken, err := session.GenerateCSRFToken()
if err != nil {
    return nil, fmt.Errorf("CSRFトークン生成エラー: %w", err)
}

// セッションデータを作成（Railsのwarden形式）
sessionData := map[string]any{
    "warden.user.user.key": []any{
        []any{userID},
        authenticatableSalt,
    },
    "_csrf_token": csrfToken, // ← 修正後：適切なトークンを設定
}
```

## 以前の仮説（参考）

### 仮説 1: Go 版がセッションを作成した後の互換性問題（正解）

Go 版がログインページなどでセッションを作成し、CSRFトークンを保存した場合：

1. Go 版がセッションに `_csrf_token` を保存（**空文字列** ← これが問題だった）
2. Rails 版がセッションから `_csrf_token` を読み取る
3. Rails 版が `csrf_meta_tags` をレンダリング時に `mask_token` を呼び出す
4. `real_csrf_token(session)` が `decode_csrf_token(session[:_csrf_token])` を呼び出す
5. デコード結果が 0 バイト（期待は 32 バイト）→ エラー

### 仮説 2: Turbo Drive による meta タグ処理の問題（不正解）

Turbo Drive 自体には問題なし。根本原因は Go 版のセッション作成にあった。

### 仮説 3: リバースプロキシによる副作用（不正解）

リバースプロキシは問題なし。根本原因は Go 版のセッション作成にあった。

## 関連ファイル

- `go/internal/usecase/create_session.go` - **問題のあったファイル（修正済み）**
- `go/internal/middleware/csrf.go` - Go 版 CSRF ミドルウェア
- `go/internal/session/csrf.go` - Go 版 CSRF トークン生成
- `go/internal/session/session.go` - Go 版セッション管理
- `go/internal/middleware/reverse_proxy.go` - リバースプロキシ
- `go/cmd/server/main.go` - ミドルウェア適用順序
- `rails/config/initializers/per_form_csrf_tokens.rb` - per_form_csrf_tokens 設定
- `rails/config/initializers/session_store.rb` - セッションストア設定
- `rails/config/application.rb` - serializer = :null 設定
- `rails/app/views/application/_head.html.erb` - csrf_meta_tags 出力
