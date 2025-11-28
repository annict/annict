# セキュリティガイドライン

このドキュメントは、Go版Annictでのセキュリティベストプラクティスを説明します。

## 基本方針

Web アプリケーションのセキュリティは**最優先事項**です。以下のガイドラインを必ず守ってください。

## CSRF（Cross-Site Request Forgery）対策

すべてのフォーム送信には**CSRF トークン**を含める必要があります。

### ミドルウェア

`internal/middleware/csrf.go` - セッションベースのCSRF保護を実装

### テンプレート実装

すべての`<form method="POST">`に hidden input でトークンを追加します。

```templ
// pages/sign_in.templ
package pages

templ SignIn(ctx context.Context, csrfToken string) {
    <form method="POST" action="/sign_in">
        // CSRFトークンを追加（必須）
        <input type="hidden" name="csrf_token" value={ csrfToken } />

        // フォームフィールド
        <input type="text" name="email_username" />
        <input type="password" name="password" />

        <button type="submit">ログイン</button>
    </form>
}
```

### ハンドラー実装

```go
// internal/handler/sign_in.go
package handler

import (
    "github.com/annict/annict/internal/middleware"
)

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // CSRFトークンをセッションから取得してテンプレートに渡す
    csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

    // テンプレートをレンダリング
    layouts.Default(ctx, meta, user, pages.SignIn(ctx, csrfToken)).Render(ctx, w)
}
```

### 重要な注意点

- **GET リクエスト**: データ取得のみ、副作用を持たない
- **POST/PUT/PATCH/DELETE**: データ変更を伴う場合は必ず CSRF トークンが必要

## XSS（Cross-Site Scripting）対策

### テンプレートの自動エスケープ

templは自動でエスケープ処理を行うため、基本的に安全です。

```templ
// ✅ 自動的にエスケープされる
<p>{ user.Comment }</p>  // <script>...</script> は &lt;script&gt; になる
```

### 注意が必要なケース

`templ.Raw()` を使う場合は、データが信頼できるソースからのものであることを確認してください。

```templ
// ⚠️ 注意: 信頼できるHTMLのみ
<div>{ templ.Raw(trustedHTMLContent) }</div>

// ❌ NG: ユーザー入力を直接Raw()で使用しない
<div>{ templ.Raw(user.Comment) }</div>
```

### ユーザー入力の扱い

すべてのユーザー入力は信頼しない前提で処理します。

```go
// ✅ Good: バリデーションを実施
func (req *CommentRequest) Validate(ctx context.Context) *session.FormErrors {
    errors := &session.FormErrors{}

    if req.Comment == "" {
        errors.AddFieldError("comment", i18n.T(ctx, "comment_required"))
    }

    // 文字数制限
    if len(req.Comment) > 1000 {
        errors.AddFieldError("comment", i18n.T(ctx, "comment_too_long"))
    }

    return errors
}
```

## SQL インジェクション対策

### プリペアドステートメント

sqlc が生成するコードは自動的にプリペアドステートメントを使用するため、安全です。

```sql
-- queries/user.sql
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;
```

```go
// 自動生成されたコードは安全（プリペアドステートメント）
user, err := queries.GetUserByEmail(ctx, email)
```

### 生SQLを書く場合

やむを得ず生 SQL を書く場合は、必ずプレースホルダー（`$1`, `$2`）を使用します。

```go
// ❌ NG: SQLインジェクションの脆弱性
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID)
db.Query(query)

// ✅ Good: プレースホルダーを使用
db.Query("SELECT * FROM users WHERE id = $1", userID)
```

## パスワード管理

### ハッシュ化

bcrypt を使用してパスワードをハッシュ化します（Rails の Devise と互換性あり）。

```go
// internal/auth/password.go
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword はパスワードをbcryptでハッシュ化
func HashPassword(password string) (string, error) {
    // コスト10を使用（Deviseのデフォルト）
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
    if err != nil {
        return "", err
    }
    return string(hashedBytes), nil
}

// CheckPassword はパスワードを検証
func CheckPassword(hashedPassword, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
```

### 重要な注意点

- **ソルト**: bcrypt が自動的に生成・管理
- **平文パスワード**: **絶対に**ログに出力しない、DB に保存しない
- **検証**: `internal/auth` パッケージの`CheckPassword`関数を使用

```go
// ❌ NG: 平文パスワードをログに出力
slog.InfoContext(ctx, "ユーザーログイン試行", "password", password)

// ✅ Good: パスワードはログに出力しない
slog.InfoContext(ctx, "ユーザーログイン試行", "email", email)
```

## 入力バリデーション

### フロントエンド・バックエンド両方で実施

```html
<!-- フロントエンド: HTML5バリデーション -->
<input type="email" name="email" required />
```

```go
// バックエンド: Request DTOでバリデーション
func (req *SignUpRequest) Validate(ctx context.Context) *session.FormErrors {
    errors := &session.FormErrors{}

    if req.Email == "" {
        errors.AddFieldError("email", i18n.T(ctx, "email_required"))
    }

    if !emailRegex.MatchString(req.Email) {
        errors.AddFieldError("email", i18n.T(ctx, "email_invalid"))
    }

    return errors
}
```

### ホワイトリスト方式

許可する値を明示的に定義します。

```go
// ✅ Good: ホワイトリスト方式
var allowedSeasons = map[string]bool{
    "spring": true,
    "summer": true,
    "autumn": true,
    "winter": true,
}

func (req *CreateWorkRequest) Validate(ctx context.Context) *session.FormErrors {
    errors := &session.FormErrors{}

    if req.Season != "" && !allowedSeasons[req.Season] {
        errors.AddFieldError("season", i18n.T(ctx, "season_invalid"))
    }

    return errors
}
```

## 認可（Authorization）

### 認証と認可は別

ログインしているだけでは不十分です。ユーザーが操作権限を持つリソースかチェックします。

```go
// internal/handler/work.go
package handler

func (h *Handler) DeleteWork(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    workID := chi.URLParam(r, "id")

    // 認証チェック: ログインしているか
    user := authMiddleware.GetUserFromContext(ctx)
    if user == nil {
        http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
        return
    }

    // 作品を取得
    work, err := h.queries.GetWorkByID(ctx, workID)
    if err != nil {
        http.NotFound(w, r)
        return
    }

    // 認可チェック: この作品の所有者か
    if work.UserID != user.ID {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // 削除処理
    // ...
}
```

### 実装例: 視聴記録の編集

視聴記録の編集・削除は、その記録の所有者のみが可能です。

```go
func (h *Handler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    recordID := chi.URLParam(r, "id")

    user := authMiddleware.GetUserFromContext(ctx)
    if user == nil {
        http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
        return
    }

    record, err := h.queries.GetRecordByID(ctx, recordID)
    if err != nil {
        http.NotFound(w, r)
        return
    }

    // 所有者チェック
    if record.UserID != user.ID {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // 削除処理
    err = h.queries.DeleteRecord(ctx, recordID)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/records", http.StatusSeeOther)
}
```

## エラーメッセージ

### 詳細な情報を漏らさない

システムの内部構造や SQL エラーをユーザーに見せないようにします。

```go
// ❌ NG: 詳細なエラーメッセージをユーザーに表示
http.Error(w, fmt.Sprintf("SQL error: %v", err), http.StatusInternalServerError)

// ✅ Good: 一般的なエラーメッセージを表示
http.Error(w, "Internal Server Error", http.StatusInternalServerError)

// サーバー側のログに詳細を記録
slog.ErrorContext(ctx, "データベースエラー", "error", err, "query", "GetUserByID")
```

### ログ出力

詳細なエラーはサーバー側のログに記録します。

```go
// ✅ Good: ログに詳細を記録
slog.ErrorContext(ctx, "パスワードリセットトークンの作成に失敗",
    "error", err,
    "user_id", userID,
)

// ユーザーには一般的なメッセージ
sessionManager.SetFlash(ctx, "alert", i18n.T(ctx, "internal_server_error"))
http.Redirect(w, r, "/", http.StatusSeeOther)
```

## セキュリティヘッダー

将来的に以下のセキュリティヘッダーの追加を検討：

```go
// 将来の実装例
func securityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        // Content-Security-Policy は慎重に設定
        next.ServeHTTP(w, r)
    })
}
```

## Cookie の設定

セッションCookieは適切な属性を設定します。

```go
// internal/middleware/session.go
cookie := &http.Cookie{
    Name:     "session_id",
    Value:    sessionID,
    Path:     "/",
    HttpOnly: true,  // JavaScriptからアクセス不可
    Secure:   true,  // HTTPS のみ（本番環境）
    SameSite: http.SameSiteLaxMode,  // CSRF 対策
    MaxAge:   86400 * 30,  // 30日間
}
http.SetCookie(w, cookie)
```

## セキュリティチェックリスト

新機能を実装する際は、以下を必ず確認してください：

### フォーム送信

- [ ] フォーム送信に CSRF トークンを含めているか
- [ ] POST/PUT/PATCH/DELETE メソッドで副作用のある処理を行っているか

### ユーザー入力

- [ ] ユーザー入力をバリデーションしているか
- [ ] ホワイトリスト方式でバリデーションしているか
- [ ] 文字数制限を設定しているか

### データベース

- [ ] SQL クエリはプリペアドステートメントを使用しているか
- [ ] sqlc で生成したコードを使用しているか

### パスワード

- [ ] パスワードは bcrypt でハッシュ化されているか
- [ ] 平文パスワードをログに出力していないか

### 認証・認可

- [ ] 認証チェックを行っているか（ログインしているか）
- [ ] 認可チェックを行っているか（操作権限があるか）
- [ ] リソースの所有者チェックを行っているか

### エラー処理

- [ ] エラーメッセージは適切か（詳細な情報を漏らしていないか）
- [ ] 詳細なエラーはログに記録しているか

### Cookie

- [ ] Cookie の設定は適切か（HttpOnly, Secure, SameSite）

## トラブルシューティング

### CSRF トークンエラー

**症状**: フォーム送信時に "Invalid CSRF Token" エラー

**原因**:
1. フォームに CSRF トークンが含まれていない
2. ミドルウェアが正しく設定されていない
3. セッションが切れている

**解決方法**:
```templ
// CSRFトークンを必ず含める
<input type="hidden" name="csrf_token" value={ csrfToken } />
```

### 認可エラー

**症状**: 403 Forbidden が表示される

**原因**:
1. 認可チェックのロジックが間違っている
2. ユーザー情報が正しく取得できていない

**解決方法**:
```go
// デバッグログを追加
slog.InfoContext(ctx, "認可チェック",
    "user_id", user.ID,
    "resource_owner_id", resource.UserID,
)
```

## 参考資料

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/)
- [Go Security Best Practices](https://github.com/OWASP/Go-SCP)
