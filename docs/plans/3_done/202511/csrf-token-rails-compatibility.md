# CSRF トークンの Rails 互換性対応 設計書

## 概要

Go 版と Rails 版で CSRF トークンの保存方法が異なるため、Go 版のページを経由して Rails 版の画像アップロードフォームを送信すると、CSRF 検証エラー（`ArgumentError: inputs must be of equal length`）が発生する問題を解決する。

**目的**:

- Go 版と Rails 版のセッション共有を正しく機能させる
- Rails 版の画像アップロードフォームなど、既存機能を正常に動作させる
- リバースプロキシによる段階的移行を円滑に進める

**背景**:

- **Go 版**: `justinas/nosurf` ライブラリを使用し、CSRF トークンを **Cookie** に保存
- **Rails 版**: CSRF トークンを **セッション（PostgreSQL）** に `_csrf_token` キーで保存
- この違いにより、Go 版でセッションを作成した後に Rails 版の保護されたフォームを送信すると、Rails 版がセッションから CSRF トークンを読み取れずエラーになる

## 要件

### 機能要件

- Go 版がセッションを作成する際、Rails 版互換の CSRF トークンをセッションに保存できる
- Rails 版の CSRF 検証が正常に動作する（既存の動作を維持）
- Go 版のページ → Rails 版のフォーム送信のフローが正常に動作する

### 非機能要件

**セキュリティ**:

- CSRF トークンは暗号学的に安全なランダム値を使用する
- トークンの長さは Rails 版と同じ（Base64 エンコードで 32 バイト）
- セッション固定攻撃を防ぐため、セッション作成時にトークンを生成

**パフォーマンス**:

- セッション読み書きのオーバーヘッドを最小限に抑える

**保守性**:

- Rails 版の CSRF トークン生成ロジックを正確に模倣する
- 将来的に Rails 版を完全に置き換える際にも問題が起きないようにする

## 設計

### 技術スタック

**Go 版**:

- `crypto/rand`: 暗号学的に安全な乱数生成
- `encoding/base64`: Base64 エンコード
- `internal/session`: セッション管理パッケージ

**Rails 版**:

- `ActionController::RequestForgeryProtection`: CSRF 保護機構
- `ActiveRecord::SessionStore`: セッションストア

### 現在の問題の詳細

#### Go 版の CSRF 実装（`internal/middleware/csrf.go`）

```go
// justinas/nosurf ライブラリを使用
csrfHandler := nosurf.New(next)
csrfHandler.SetBaseCookie(...) // Cookieに保存
```

- **保存先**: Cookie（`_csrf_token` という名前の Cookie）
- **セッションには保存しない**

#### Rails 版の CSRF 実装

```ruby
# config/initializers/session_store.rb
Annict::Application.config.session_store :active_record_store,
  key: "_annict_session_v201904",
  domain: ENV["ANNICT_COOKIE_DOMAIN"].presence,
  expire_after: 30.days
```

- **保存先**: PostgreSQL の `sessions` テーブル（JSONB の `data` カラム）
- **キー名**: `_csrf_token`

#### 問題の発生フロー

1. **Go 版のページ（例: `/sign_in`）にアクセス**
   - Go 版が nosurf を使って CSRF トークンを **Cookie** に保存
   - **セッションには CSRF トークンが保存されない**

2. **Rails 版の画像アップロードフォームを表示**
   - Rails 版がセッションから CSRF トークンを読み取ろうとする
   - しかし、**セッションに CSRF トークンがない**
   - Rails 版が新しい CSRF トークンを生成してフォームに埋め込む

3. **画像アップロードフォームを送信**
   - フォームに含まれる CSRF トークン（Rails 版が生成したもの）
   - セッションに保存されている CSRF トークン（`nil` または空文字列）
   - **2 つのトークンの長さが一致せず**、`ArgumentError: inputs must be of equal length` が発生

### 解決方針

**オプション 1: Go 版のミドルウェアで Rails 互換の CSRF トークンをセッションに保存する（採用）**

- Go 版が新規セッションを作成する際に、Rails 版と同じ形式の CSRF トークンをセッションに保存
- `justinas/nosurf` は使用せず、独自の CSRF 保護ミドルウェアを実装
- Rails 版のトークン生成ロジック（`SecureRandom.base64(32)`）を模倣

**メリット**:

- Rails 版の変更が不要
- Rails 版との完全な互換性を保てる
- セッション管理が一元化される

**デメリット**:

- Go 版で独自の CSRF 保護ミドルウェアを実装する必要がある

### コード設計

#### Go 版: セッション作成時の CSRF トークン生成

```go
// internal/session/csrf.go (新規作成)
package session

import (
    "crypto/rand"
    "encoding/base64"
)

// GenerateCSRFToken はRails互換のCSRFトークンを生成
// Rails版: SecureRandom.base64(32)
func GenerateCSRFToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(b), nil
}
```

#### Go 版: セッション作成時に CSRF トークンを保存

```go
// internal/session/session.go
func (m *Manager) CreateSession(ctx context.Context, w http.ResponseWriter, userID int64) error {
    // 1. Public IDを生成
    publicID, err := generatePublicID()
    if err != nil {
        return err
    }

    // 2. CSRFトークンを生成
    csrfToken, err := GenerateCSRFToken()
    if err != nil {
        return err
    }

    // 3. セッションデータを作成
    sessionData := map[string]any{
        "warden.user.user.key": []any{[]any{float64(userID)}, "authenticatable_salt"},
        "_csrf_token":          csrfToken,
    }

    // 4. JSONにエンコード
    jsonData, err := json.Marshal(sessionData)
    if err != nil {
        return err
    }

    // 5. DBに保存
    _, err = m.sessionRepo.CreateSession(ctx, publicID, jsonData)
    if err != nil {
        return err
    }

    // 6. Cookieを設定
    cookie := &http.Cookie{
        Name:     SessionKey,
        Value:    publicID,
        Path:     "/",
        Domain:   m.cfg.CookieDomain,
        Secure:   m.cfg.SessionSecure == "true",
        HttpOnly: m.cfg.SessionHTTPOnly == "true",
        SameSite: http.SameSiteLaxMode,
        MaxAge:   30 * 24 * 60 * 60, // 30日
    }
    http.SetCookie(w, cookie)

    return nil
}
```

#### Go 版: CSRF 保護ミドルウェアの実装

```go
// internal/middleware/csrf.go (書き換え)
package middleware

import (
    "net/http"

    "github.com/annict/annict/internal/session"
)

// CSRFMiddleware はCSRF保護ミドルウェア
type CSRFMiddleware struct {
    sessionManager *session.Manager
}

func NewCSRFMiddleware(sessionManager *session.Manager) *CSRFMiddleware {
    return &CSRFMiddleware{
        sessionManager: sessionManager,
    }
}

func (m *CSRFMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // GETリクエストはCSRFチェックをスキップ
        if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
            next.ServeHTTP(w, r)
            return
        }

        // セッションからCSRFトークンを取得
        sessionID, err := m.sessionManager.GetSessionID(r)
        if err != nil || sessionID == "" {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }

        sessionData, err := m.sessionManager.GetSession(ctx, sessionID)
        if err != nil || sessionData == nil {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }

        // フォームからCSRFトークンを取得
        formToken := r.FormValue("csrf_token")
        if formToken == "" {
            formToken = r.Header.Get("X-CSRF-Token")
        }

        // トークンを比較
        if formToken != sessionData.CSRFToken {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }

        next.ServeHTTP(w, r)
    })
}

// GetCSRFToken はリクエストからCSRFトークンを取得
func GetCSRFToken(r *http.Request, sessionManager *session.Manager) string {
    ctx := r.Context()
    sessionID, err := sessionManager.GetSessionID(r)
    if err != nil || sessionID == "" {
        return ""
    }

    sessionData, err := sessionManager.GetSession(ctx, sessionID)
    if err != nil || sessionData == nil {
        return ""
    }

    return sessionData.CSRFToken
}
```

### テスト戦略

**Go 版 - 単体テスト**:

- `GenerateCSRFToken()` が Base64 エンコードされた 32 バイトのランダム文字列を生成することをテスト
- セッション作成時に CSRF トークンがセッションに保存されることをテスト
- CSRF ミドルウェアが正しくトークンを検証することをテスト

**Go 版 - 統合テスト**:

- Go 版のログインページ → Rails 版の画像アップロードフォーム送信のフローをテスト
- セッションと CSRF トークンが正しく共有されることを確認

**Rails 版 - 手動テスト**:

- Go 版でログイン → Rails 版の画像アップロードフォームを送信
- エラーが発生しないことを確認

## 調査結果（フェーズ 1）

### タスク 1-1: Rails 版の CSRF トークン生成・検証ロジックの調査結果

#### Rails 版の CSRF 保護設定

**セッションストア設定** (`config/initializers/session_store.rb`):

```ruby
Annict::Application.config.session_store :active_record_store,
  key: "_annict_session_v201904",
  domain: ENV["ANNICT_COOKIE_DOMAIN"].presence,
  expire_after: 30.days
```

**CSRF 保護の有効化** (`app/controllers/application_controller.rb`):

```ruby
protect_from_forgery with: :exception, prepend: true
```

#### CSRF トークンの生成方法

Rails 版の CSRF トークン生成は、`ActionController::RequestForgeryProtection`モジュールで実装されています。

**トークン生成コード**（Rails ソースコードより）:

```ruby
session[:_csrf_token] ||= SecureRandom.base64(32)
```

**生成仕様**:

- **メソッド**: `SecureRandom.base64(32)`
- **入力**: 32 バイトのランダムデータ
- **出力**: Base64 エンコードされた文字列（約 44 文字）
- **暗号学的安全性**: `SecureRandom`は暗号学的に安全な乱数生成器を使用

#### セッションへの保存方法

**キー名**: `_csrf_token`（シンボル `:_csrf_token` または文字列 `"_csrf_token"`）

**保存先**: PostgreSQL の`sessions`テーブル

**テーブル構造**:

```sql
CREATE TABLE public.sessions (
    id bigint NOT NULL,
    session_id character varying NOT NULL,
    data jsonb DEFAULT '"{}"'::jsonb NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);
```

**データ形式**: `data`カラムに JSONB 形式で保存

#### 実際のセッションデータ（PostgreSQL 確認結果）

開発環境の PostgreSQL データベースで実際のセッションデータを確認しました：

```sql
SELECT id, session_id, data, created_at, updated_at
FROM sessions
ORDER BY updated_at DESC
LIMIT 3;
```

**結果例**:

```jsonb
{
  "flash": "{\"type\":\"success\",\"message\":\"ユーザー登録が完了しました。Annictへようこそ！\"}",
  "_csrf_token": "",
  "user_return_to": "/"
}
```

```jsonb
{
  "_csrf_token": "",
  "warden.user.user.key": [[320394], ""]
}
```

**重要な発見**:

- 現在のセッションデータでは `_csrf_token` が**空文字列**になっている
- これは、Go 版がセッションを作成した際に`_csrf_token`キーを保存していないため
- Rails 版はセッションに`_csrf_token`が存在しない場合、新しいトークンを生成する
- この動作の不一致が、Go 版 →Rails 版のフロー時に CSRF 検証エラーを引き起こす

#### Go 版で実装すべき内容（まとめ）

1. **CSRF トークン生成関数**:
   - `crypto/rand`で 32 バイトのランダムデータを生成
   - `encoding/base64.StdEncoding.EncodeToString()`で Base64 エンコード
   - Rails 版の`SecureRandom.base64(32)`と同等の処理

2. **セッション保存**:
   - セッション作成時に`_csrf_token`キーで JSONB データに保存
   - キー名は文字列の`"_csrf_token"`を使用

3. **互換性の確保**:
   - トークンの長さ（Base64 エンコード後：約 44 文字）
   - トークンの形式（Base64 エンコード）
   - セッションのキー名（`_csrf_token`）

### タスク 1-2: Go 版の現在の CSRF 実装の調査結果

#### Go 版の CSRF 実装の現状

Go 版は現在、`justinas/nosurf` ライブラリを使用して CSRF 保護を実装しています。Rails 版とは異なり、CSRF トークンを**Cookie**に保存しており、**セッション（PostgreSQL）には保存していません**。

#### `internal/middleware/csrf.go` の実装

**現在の実装** (`internal/middleware/csrf.go:10-27`):

```go
// NewCSRFMiddleware はCSRF保護ミドルウェアを作成
func NewCSRFMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		csrfHandler := nosurf.New(next)

		// CSRFトークンのCookie設定
		csrfHandler.SetBaseCookie(http.Cookie{
			Path:     "/",
			Domain:   cfg.CookieDomain,
			MaxAge:   nosurf.MaxAge,
			Secure:   cfg.SessionSecure == "true",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		return csrfHandler
	}
}
```

**トークン取得** (`internal/middleware/csrf.go:29-33`):

```go
// GetCSRFToken はリクエストからCSRFトークンを取得
// テンプレートでトークンを表示する際に使用
func GetCSRFToken(r *http.Request) string {
	return nosurf.Token(r)
}
```

**重要な発見**:

- `justinas/nosurf` を使用して CSRF 保護を実装
- CSRF トークンを**Cookie**に保存（`SetBaseCookie()`）
- **セッション（PostgreSQL）には保存されない**
- トークン取得は `nosurf.Token(r)` で行う

#### `internal/session/session.go` の実装

**SessionData 構造体** (`internal/session/session.go:37-42`):

```go
// SessionData はセッションに保存されているデータ
type SessionData struct {
	UserID    *int64         `json:"warden.user.user.key"`
	CSRFToken string         `json:"_csrf_token"`
	ExtraData map[string]any `json:"-"`
}
```

**GetSession()メソッド** (`internal/session/session.go:60-106`):

```go
// GetSession はセッションIDからセッションデータを取得
func (m *Manager) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	// ... (中略) ...

	// CSRFトークンを取得
	if csrfToken, ok := rawData["_csrf_token"].(string); ok {
		sessionData.CSRFToken = csrfToken
	}

	return sessionData, nil
}
```

**SetValue()メソッド** (`internal/session/session.go:140-215`):

セッションに任意のキー・バリューを保存できるメソッドが存在します。新規セッションを作成する場合は、`generatePublicID()` でランダムな public ID を生成し、Cookie を設定します。

**generatePublicID()** (`internal/session/session.go:347-355`):

```go
// generatePublicID ランダムなpublic IDを生成
func generatePublicID() (string, error) {
	// 32バイトのランダムデータを生成
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}
```

**重要な発見**:

- ✅ `SessionData`構造体に`CSRFToken`フィールドは定義済み
- ✅ `GetSession()`でセッションから CSRF トークンを**読み取る**処理は実装済み
- ❌ セッション作成時に CSRF トークンを**生成・保存する**処理がない
- ❌ Rails 互換の CSRF トークン生成関数（`SecureRandom.base64(32)`相当）がない

#### 問題点の整理

| 項目                 | Go 版（現状）     | Rails 版                        | 互換性 |
| -------------------- | ----------------- | ------------------------------- | ------ |
| **CSRF 保存先**      | Cookie（nosurf）  | セッション（PostgreSQL）        | ❌     |
| **セッションに保存** | なし              | `_csrf_token`キー               | ❌     |
| **トークン生成方法** | nosurf が自動生成 | `SecureRandom.base64(32)`       | ❌     |
| **トークン形式**     | nosurf が決定     | Base64 エンコード（約 44 文字） | ❌     |

**問題の発生フロー**（再確認）:

1. **Go 版のページにアクセス** → CSRF トークンを**Cookie**に保存（セッションには保存しない）
2. **Rails 版のフォームを表示** → セッションに CSRF トークンがないため、Rails 版が新しいトークンを生成
3. **フォーム送信** → フォームのトークン（Rails 版が生成）とセッションのトークン（nil）が一致せず、`ArgumentError: inputs must be of equal length` が発生

#### Go 版で実装すべき内容（まとめ）

1. **CSRF トークン生成関数** (`internal/session/csrf.go` を新規作成):
   - `crypto/rand`で 32 バイトのランダムデータを生成
   - `encoding/base64.StdEncoding.EncodeToString()`で Base64 エンコード
   - Rails 版の`SecureRandom.base64(32)`と同等の処理

2. **セッション作成時の CSRF トークン保存** (`internal/session/session.go`):
   - 新規セッション作成時に自動的に CSRF トークンを生成・保存
   - `SetValue()`メソッドを活用するか、専用のロジックを追加
   - キー名は文字列の`"_csrf_token"`を使用

3. **CSRF 保護ミドルウェアの書き換え** (`internal/middleware/csrf.go`):
   - `justinas/nosurf`を削除
   - セッションから CSRF トークンを読み取り、フォームのトークンと比較
   - GET リクエストは CSRF チェックをスキップ
   - `GetCSRFToken()`関数をセッションベースに変更

4. **互換性の確保**:
   - トークンの長さ（Base64 エンコード後：約 44 文字）
   - トークンの形式（Base64 エンコード）
   - セッションのキー名（`_csrf_token`）

## タスクリスト

### フェーズ 1: 調査

- [x] **1-1**: Rails 版の CSRF トークン生成・検証ロジックを調査（Rails 版コンテナ）
  - Rails 版のソースコード（`ActionController::RequestForgeryProtection`）を確認
  - CSRF トークンの形式（Base64 エンコード、長さ）を確認
  - セッションへの保存方法（`_csrf_token` キー）を確認
  - 実際のセッションデータを PostgreSQL で確認
  - **想定ファイル数**: 0 ファイル（調査のみ）
  - **想定行数**: 0 行

- [x] **1-2**: Go 版の現在の CSRF 実装を調査（Go 版コンテナ）
  - `internal/middleware/csrf.go` の実装を確認
  - `justinas/nosurf` の動作を確認
  - セッション作成時の処理を確認（`internal/session/session.go`）
  - **想定ファイル数**: 0 ファイル（調査のみ）
  - **想定行数**: 0 行

### フェーズ 2: Go 版の実装

- [x] **2-1**: CSRF トークン生成関数の実装（Go 版コンテナ）
  - `internal/session/csrf.go` を新規作成
  - `GenerateCSRFToken()` 関数を実装（Rails 互換）
  - 単体テストを作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 100 行（実装 30 行 + テスト 70 行）

- [x] **2-2**: セッション作成時の CSRF トークン保存処理を実装（Go 版コンテナ）
  - `internal/session/session.go` の `CreateSession()` メソッドを新規作成または修正
  - セッション作成時に `_csrf_token` をセッションに保存
  - 既存の `SetValue()` メソッドも必要に応じて修正
  - 単体テストを作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

- [x] **2-3**: CSRF 保護ミドルウェアの実装（Go 版コンテナ）
  - `internal/middleware/csrf.go` を書き換え
  - `justinas/nosurf` を削除し、独自の CSRF 保護ミドルウェアを実装
  - セッションから CSRF トークンを読み取り、フォームのトークンと比較
  - 単体テストを作成
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [x] **2-4**: テンプレートでの CSRF トークン取得方法を修正（Go 版コンテナ）
  - `internal/middleware/csrf.go` の `GetCSRFToken()` 関数を修正
  - テンプレート（`.templ` ファイル）での CSRF トークン取得を修正
  - ハンドラーでの使用方法を修正
  - **想定ファイル数**: 約 5 ファイル（実装 5 + テスト 0）
  - **想定行数**: 約 50 行（実装 50 行 + テスト 0 行）

### フェーズ 3: テストと検証

- [x] **3-1**: Go 版の統合テストを実装（Go 版コンテナ）
  - セッション作成 → CSRF トークンが保存されることをテスト
  - フォーム送信 → CSRF 検証が正常に動作することをテスト
  - 不正なトークン → 403 エラーが返されることをテスト
  - **想定ファイル数**: 約 1 ファイル（実装 0 + テスト 1）
  - **想定行数**: 約 150 行（実装 0 行 + テスト 150 行）

- [x] **3-2**: 初回アクセス時の CSRF トークン生成バグを修正（Go 版コンテナ）
  - **問題**: 初めてログインページにアクセスしたユーザーはセッションが存在せず、CSRF トークンが空のままフォーム送信されて 403 エラーが発生
  - **修正内容**:
    - `internal/middleware/csrf.go` に `GetOrCreateCSRFToken()` 関数を追加
    - セッションが存在しない場合は自動的に新規セッションを作成し、CSRF トークンを生成
    - テスト環境では `Result()` メソッドで Cookie を取得し、新しいセッション ID を使って CSRF トークンを取得
    - ログイン前のページ（`sign_in/new.go`, `sign_up/new.go`, `password_reset/new.go`）で `GetOrCreateCSRFToken()` を使用
  - テストの追加:
    - `TestGetOrCreateCSRFToken_NoSession`: セッションがない場合、新規セッションを作成して CSRF トークンを生成することを確認
    - `TestGetOrCreateCSRFToken_ExistingSession`: セッションが既に存在する場合、既存の CSRF トークンを返すことを確認
  - 既存のテストで動作確認（CSRF ミドルウェア、ログイン、サインアップ、パスワードリセット）
  - **想定ファイル数**: 5 ファイル（実装 4 + テスト 1）
  - **想定行数**: 約 150 行（実装 50 行 + テスト 100 行）

- [ ] **3-3**: `justinas/nosurf` の依存関係を削除（Go 版コンテナ）
  - `go.mod` から `justinas/nosurf` を削除
  - `go mod tidy` を実行
  - **想定ファイル数**: 約 2 ファイル（go.mod, go.sum）
  - **想定行数**: 約 5 行

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **Rails 版の CSRF 保護の変更**: Rails 版は既存の実装をそのまま使用する
- **Go 版での CSRF トークンのローテーション**: 現時点ではセッション作成時に 1 度だけ生成する
- **Per-form CSRF tokens**: Rails 版は `per_form_csrf_tokens = true` だが、Go 版では当面対応しない

## 参考資料

- [Rails - ActionController::RequestForgeryProtection](https://api.rubyonrails.org/classes/ActionController/RequestForgeryProtection.html)
- [justinas/nosurf](https://github.com/justinas/nosurf)
- [OWASP - Cross-Site Request Forgery (CSRF)](https://owasp.org/www-community/attacks/csrf)
- [Rails - ActiveRecord::SessionStore](https://github.com/rails/activerecord-session_store)
