# リクエストバリデーションガイド

このドキュメントは、Go版Annictでのリクエストバリデーションのベストプラクティスを説明します。

## 概要

フォームからの入力値の検証は、**Request DTO（Data Transfer Object）パターン**を使用します。

### 基本方針

- **責務**: リクエストデータの構造定義と**形式バリデーションのみ**を行う
- **ファイル配置**: ハンドラーと同じディレクトリ（例: `internal/handler/sign_in_request.go`）
- **命名規則**: `{Action}Request` （例: `SignInRequest`, `CreateWorkRequest`）

### バリデーション範囲

- ✅ **形式チェック**: 必須チェック、フォーマット検証、文字数制限など（DB アクセス不要）
- ❌ **ビジネスロジック**: ユーザー存在チェック、パスワード照合など（DB アクセス必要）→ ハンドラーで実行

### メッセージの国際化

バリデーションメッセージは必ず `i18n.T(ctx, "message_id")` を使用します。

## 実装例

### シンプルなバリデーション

```go
// internal/handler/sign_in_request.go
package handler

import (
    "context"
    "github.com/annict/annict/internal/i18n"
    "github.com/annict/annict/internal/session"
)

type SignInRequest struct {
    EmailOrUsername string
    Password        string
}

// 形式バリデーションのみ（DBアクセスなし）
func (req *SignInRequest) Validate(ctx context.Context) *session.FormErrors {
    errors := &session.FormErrors{}

    if req.EmailOrUsername == "" {
        errors.AddFieldError("email_username", i18n.T(ctx, "sign_in_error_email_username_required"))
    }

    if req.Password == "" {
        errors.AddFieldError("password", i18n.T(ctx, "sign_in_error_password_required"))
    }

    if errors.HasErrors() {
        return errors
    }
    return nil
}
```

### 複雑なバリデーション

```go
// internal/handler/password_reset_request.go
package handler

import (
    "context"
    "github.com/annict/annict/internal/i18n"
    "github.com/annict/annict/internal/session"
    "regexp"
)

type PasswordResetRequest struct {
    Email string
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (req *PasswordResetRequest) Validate(ctx context.Context) *session.FormErrors {
    errors := &session.FormErrors{}

    // 必須チェック
    if req.Email == "" {
        errors.AddFieldError("email", i18n.T(ctx, "password_reset_email_required"))
        return errors // 後続のチェックをスキップ
    }

    // フォーマットチェック
    if !emailRegex.MatchString(req.Email) {
        errors.AddFieldError("email", i18n.T(ctx, "password_reset_email_invalid"))
    }

    // 文字数制限
    if len(req.Email) > 255 {
        errors.AddFieldError("email", i18n.T(ctx, "password_reset_email_too_long"))
    }

    if errors.HasErrors() {
        return errors
    }
    return nil
}
```

### 複数フィールドのバリデーション

```go
// internal/handler/update_password_request.go
package handler

import (
    "context"
    "github.com/annict/annict/internal/i18n"
    "github.com/annict/annict/internal/session"
)

type UpdatePasswordRequest struct {
    Password             string
    PasswordConfirmation string
}

func (req *UpdatePasswordRequest) Validate(ctx context.Context) *session.FormErrors {
    errors := &session.FormErrors{}

    // 必須チェック
    if req.Password == "" {
        errors.AddFieldError("password", i18n.T(ctx, "update_password_password_required"))
    }

    if req.PasswordConfirmation == "" {
        errors.AddFieldError("password_confirmation", i18n.T(ctx, "update_password_password_confirmation_required"))
    }

    // 文字数チェック
    if len(req.Password) < 8 {
        errors.AddFieldError("password", i18n.T(ctx, "update_password_password_too_short"))
    }

    // パスワード一致チェック
    if req.Password != "" && req.PasswordConfirmation != "" && req.Password != req.PasswordConfirmation {
        errors.AddFieldError("password_confirmation", i18n.T(ctx, "update_password_password_mismatch"))
    }

    if errors.HasErrors() {
        return errors
    }
    return nil
}
```

## ハンドラーでの使用

### 基本パターン

```go
func (h *Handler) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // リクエストを作成
    req := &SignInRequest{
        EmailOrUsername: r.FormValue("email_username"),
        Password:        r.FormValue("password"),
    }

    // 形式バリデーション
    if formErrors := req.Validate(ctx); formErrors != nil {
        // エラーをセッションに保存
        sessionManager := session.GetSessionManager(r)
        sessionManager.SetFormErrors(ctx, formErrors)

        // フォームを再表示
        http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
        return
    }

    // ビジネスロジック（DB検索、パスワード照合など）
    user, err := h.queries.GetUserByEmailOrUsername(ctx, req.EmailOrUsername)
    if err != nil {
        // ユーザーが見つからない場合のエラー処理
        formErrors := &session.FormErrors{}
        formErrors.AddGeneralError(i18n.T(ctx, "sign_in_error_invalid_credentials"))
        sessionManager.SetFormErrors(ctx, formErrors)
        http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
        return
    }

    // パスワード照合
    if err := auth.CheckPassword(user.EncryptedPassword, req.Password); err != nil {
        // パスワードが一致しない場合のエラー処理
        formErrors := &session.FormErrors{}
        formErrors.AddGeneralError(i18n.T(ctx, "sign_in_error_invalid_credentials"))
        sessionManager.SetFormErrors(ctx, formErrors)
        http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
        return
    }

    // 認証成功
    // ...
}
```

### PATCHメソッドでの使用（Method Override）

```go
func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Method Overrideミドルウェアにより、POSTリクエストがPATCHに変換される
    // フォーム: <input type="hidden" name="_method" value="PATCH" />

    req := &UpdatePasswordRequest{
        Password:             r.FormValue("password"),
        PasswordConfirmation: r.FormValue("password_confirmation"),
    }

    // 形式バリデーション
    if formErrors := req.Validate(ctx); formErrors != nil {
        sessionManager := session.GetSessionManager(r)
        sessionManager.SetFormErrors(ctx, formErrors)
        http.Redirect(w, r, "/password/edit", http.StatusSeeOther)
        return
    }

    // ビジネスロジック
    // ...
}
```

## テスト

### Request DTOのテスト

```go
func TestSignInRequest_Validate(t *testing.T) {
    tests := []struct {
        name          string
        request       SignInRequest
        wantErrors    bool
        expectedField string
    }{
        {
            name: "valid request",
            request: SignInRequest{
                EmailOrUsername: "user@example.com",
                Password:        "password123",
            },
            wantErrors: false,
        },
        {
            name: "missing email_username",
            request: SignInRequest{
                EmailOrUsername: "",
                Password:        "password123",
            },
            wantErrors:    true,
            expectedField: "email_username",
        },
        {
            name: "missing password",
            request: SignInRequest{
                EmailOrUsername: "user@example.com",
                Password:        "",
            },
            wantErrors:    true,
            expectedField: "password",
        },
        {
            name: "missing both fields",
            request: SignInRequest{
                EmailOrUsername: "",
                Password:        "",
            },
            wantErrors: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            ctx = i18n.WithLocale(ctx, "ja")

            formErrors := tt.request.Validate(ctx)

            if tt.wantErrors {
                if formErrors == nil {
                    t.Error("expected errors, got nil")
                    return
                }
                if !formErrors.HasErrors() {
                    t.Error("expected errors, got none")
                }
                if tt.expectedField != "" && !formErrors.HasFieldError(tt.expectedField) {
                    t.Errorf("expected field error for %q", tt.expectedField)
                }
            } else {
                if formErrors != nil {
                    t.Errorf("expected no errors, got %v", formErrors)
                }
            }
        })
    }
}
```

### 正規表現のテスト

```go
func TestEmailRegex(t *testing.T) {
    tests := []struct {
        email string
        valid bool
    }{
        {"user@example.com", true},
        {"user.name@example.co.jp", true},
        {"user+tag@example.com", true},
        {"", false},
        {"invalid", false},
        {"@example.com", false},
        {"user@", false},
        {"user@.com", false},
    }

    for _, tt := range tests {
        t.Run(tt.email, func(t *testing.T) {
            if got := emailRegex.MatchString(tt.email); got != tt.valid {
                t.Errorf("emailRegex.MatchString(%q) = %v, want %v", tt.email, got, tt.valid)
            }
        })
    }
}
```

## ベストプラクティス

### 1. 形式チェックとビジネスロジックを分離

```go
// ❌ Bad: Request DTOでDBアクセス
func (req *SignInRequest) Validate(ctx context.Context, queries *repository.Queries) *session.FormErrors {
    // ...
    user, err := queries.GetUserByEmail(ctx, req.Email)
    if err != nil {
        // ユーザー存在チェック
    }
    // ...
}

// ✅ Good: Request DTOは形式チェックのみ
func (req *SignInRequest) Validate(ctx context.Context) *session.FormErrors {
    // 必須チェック、フォーマット検証のみ
}

// ハンドラーでビジネスロジックを実行
func (h *Handler) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
    // 形式チェック
    if formErrors := req.Validate(ctx); formErrors != nil {
        // ...
    }

    // ビジネスロジック（ユーザー存在チェック、パスワード照合など）
    user, err := h.queries.GetUserByEmail(ctx, req.Email)
    // ...
}
```

### 2. 国際化を徹底

```go
// ❌ Bad: ハードコードされたメッセージ
errors.AddFieldError("email", "メールアドレスを入力してください")

// ✅ Good: 国際化された翻訳
errors.AddFieldError("email", i18n.T(ctx, "password_reset_email_required"))
```

### 3. 早期リターンでネストを減らす

```go
// ❌ Bad: ネストが深い
func (req *SignInRequest) Validate(ctx context.Context) *session.FormErrors {
    errors := &session.FormErrors{}
    if req.Email != "" {
        if emailRegex.MatchString(req.Email) {
            // OK
        } else {
            errors.AddFieldError("email", "...")
        }
    } else {
        errors.AddFieldError("email", "...")
    }
    return errors
}

// ✅ Good: 早期リターンでシンプル
func (req *SignInRequest) Validate(ctx context.Context) *session.FormErrors {
    errors := &session.FormErrors{}

    if req.Email == "" {
        errors.AddFieldError("email", i18n.T(ctx, "email_required"))
        return errors // 後続のチェックをスキップ
    }

    if !emailRegex.MatchString(req.Email) {
        errors.AddFieldError("email", i18n.T(ctx, "email_invalid"))
    }

    if errors.HasErrors() {
        return errors
    }
    return nil
}
```

### 4. 正規表現はパッケージレベルで定義

```go
// ✅ Good: 正規表現のコンパイルを1回だけ実行
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (req *PasswordResetRequest) Validate(ctx context.Context) *session.FormErrors {
    // emailRegexを使用
}
```

## 利点

1. **単一責任の原則**: Request は「データの形式」のみを検証
2. **テストしやすい**: DB モックなしでバリデーションをテストできる
3. **依存が少ない**: `repository.Queries` に依存しない
4. **ハンドラーの見通しが良くなる**: ハンドラーはHTTP処理とビジネスロジックに専念
5. **再利用可能**: 同じバリデーションロジックを複数のハンドラーで使用可能
