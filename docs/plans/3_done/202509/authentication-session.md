# 認証とセッション管理

## 概要

Rails 側と共有する PostgreSQL の sessions テーブルを使用したセッション管理と、
ログインユーザーの認証状態の管理機能を実装。

## 実装内容

### フェーズ 3: 認証とセッション

- [x] **Rails セッション読み取り機能**
  - [x] セッション Cookie の確認
  - [x] PostgreSQL の sessions テーブルからセッションデータ取得
  - [x] user_id の取得

- [x] **認証状態の表示**
  - [x] ログインユーザー名表示
  - [x] ログイン/ログアウトリンク

- [x] **認証が必要なページの保護**
  - [x] ミドルウェアでのチェック
  - [x] リダイレクト処理

## セッション共有の仕組み

### Rails との互換性

Go 版は Rails 側が管理する sessions テーブルを**読み取り専用**で使用します：

- **セッションストア**: PostgreSQL の `sessions` テーブル
- **セッションキー**: `_annict_session_v201904`（Rails と同一）
- **Cookie 設定**:
  - Domain: `.example.dev`（開発）、`.annict.com`（本番）
  - Secure: true（HTTPS のみ）
  - HttpOnly: true
  - SameSite: Lax

### セッションデータの形式

Rails の activerecord-session_store が作成するセッションデータ：

- `sessions.session_id`: セッション ID（Cookie の値）
- `sessions.data`: Base64 エンコードされた Marshal 形式のデータ
- `sessions.created_at`: セッション作成日時
- `sessions.updated_at`: セッション更新日時

## 実装詳細

### セッションミドルウェア

```go
// internal/middleware/session.go
type SessionMiddleware struct {
    queries *repository.Queries
}

func (m *SessionMiddleware) LoadSession(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // CookieからセッションIDを取得
        cookie, err := r.Cookie("_annict_session_v201904")
        if err != nil {
            next.ServeHTTP(w, r)
            return
        }

        // PostgreSQLからセッションデータを取得
        session, err := m.queries.GetSessionBySessionID(ctx, cookie.Value)
        if err != nil {
            next.ServeHTTP(w, r)
            return
        }

        // セッションデータをパース（user_idを抽出）
        userID, err := parseSessionData(session.Data)
        if err != nil {
            next.ServeHTTP(w, r)
            return
        }

        // ユーザー情報を取得
        user, err := m.queries.GetUserByID(ctx, userID)
        if err != nil {
            next.ServeHTTP(w, r)
            return
        }

        // コンテキストにユーザー情報を格納
        ctx = session.SetCurrentUser(ctx, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### セッションデータのパース

```go
// pkg/session/parser.go
func parseSessionData(data string) (int64, error) {
    // Base64デコード
    decoded, err := base64.StdEncoding.DecodeString(data)
    if err != nil {
        return 0, err
    }

    // Marshalデータから"warden.user.user.key"を探す
    // 簡易的な実装（正規表現やバイトパターンマッチング）
    userID := extractUserIDFromMarshal(decoded)
    return userID, nil
}
```

### 認証が必要なページの保護

```go
// internal/middleware/auth.go
func RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // セッションからユーザー情報を取得
        user := session.GetCurrentUser(ctx)
        if user == nil {
            // 未認証の場合はログインページにリダイレクト
            http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### テンプレートでの認証状態表示

```html
<!-- layouts/base.html -->
<header>
  {{if .CurrentUser}}
  <span>{{.CurrentUser.Username}}</span>
  <a href="/sign_out">ログアウト</a>
  {{else}}
  <a href="/sign_in">ログイン</a>
  {{end}}
</header>
```

## セキュリティ考慮事項

### Cookie 設定

```go
http.SetCookie(w, &http.Cookie{
    Name:     "_annict_session_v201904",
    Value:    sessionID,
    Path:     "/",
    Domain:   ".example.dev",  // Rails と共有
    Secure:   true,                // HTTPS のみ
    HttpOnly: true,                // JavaScript からアクセス不可
    SameSite: http.SameSiteLaxMode, // CSRF 対策
    MaxAge:   86400 * 30,          // 30日間
})
```

### セッション有効期限

- セッションの有効期限は Rails 側で管理
- Go 側では期限切れセッションを検出してログインページにリダイレクト

## テスト

### セッションミドルウェアのテスト

```go
func TestLoadSession(t *testing.T) {
    db, tx := testutil.SetupTestDB(t)

    // テストユーザーとセッションを作成
    userID := testutil.NewUserBuilder(t, tx).
        WithEmail("test@example.com").
        Build()

    sessionID := testutil.NewSessionBuilder(t, tx).
        WithUserID(userID).
        Build()

    // リクエストにセッションCookieを設定
    req := httptest.NewRequest("GET", "/", nil)
    req.AddCookie(&http.Cookie{
        Name:  "_annict_session_v201904",
        Value: sessionID,
    })

    // ミドルウェア実行
    middleware := NewSessionMiddleware(repository.New(db).WithTx(tx))
    handler := middleware.LoadSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := session.GetCurrentUser(r.Context())
        assert.NotNil(t, user)
        assert.Equal(t, "test@example.com", user.Email)
    }))

    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)
}
```

### 認証保護のテスト

```go
func TestRequireAuth(t *testing.T) {
    // 未認証の場合
    req := httptest.NewRequest("GET", "/protected", nil)
    rr := httptest.NewRecorder()

    handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    handler.ServeHTTP(rr, req)

    // ログインページにリダイレクトされる
    assert.Equal(t, http.StatusSeeOther, rr.Code)
    assert.Equal(t, "/sign_in", rr.Header().Get("Location"))
}
```

## 制限事項

### 読み取り専用

Go 版のセッション管理は**読み取り専用**です：

- ✅ 既存セッションの読み取り（ユーザー認証状態の確認）
- ✅ ログインユーザー情報の取得
- ❌ セッションデータの更新（Rails 側で管理）
- ❌ セッションの削除（Rails 側で管理）

### セッション作成

新しいセッションの作成は、ログイン機能実装時に Go 側でも実装：

- Rails 互換のセッションデータ形式で作成
- sessions テーブルへの INSERT
- セッション Cookie の発行

## 成果

- **Rails とのシームレスな統合**: Rails 側のログイン状態を Go 側でも認識できる
- **セキュアなセッション管理**: Cookie の適切な設定で XSS/CSRF 対策
- **ミドルウェアパターンの確立**: 認証チェックを簡潔に実装
- **テスト可能な構造**: セッション管理のテストパターンを確立

## 関連ドキュメント

- [プロジェクト全体の設計書](./go.md)
- [ログインページ](./sign-in.md)
- [テストインフラ](./testing-infrastructure.md)
