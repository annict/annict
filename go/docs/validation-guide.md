# バリデーションガイド

このドキュメントは、Go版Annictでのバリデーションのベストプラクティスを説明します。

## 概要

フォームからの入力値の検証は、**形式バリデーション**と**状態バリデーション**に分けて実装します。

### ファイル構成

```
internal/handler/sign_in/
├── handler.go              # Handler構造体と依存性
├── format_validator.go     # 形式バリデーション（入力値の形式チェック）
├── state_validator.go      # 状態バリデーション（DBを使った検証）
├── format_validator_test.go
├── state_validator_test.go
├── new.go                  # フォーム表示
└── create.go               # 作成処理
```

### バリデーションの分類

| 種類               | 責務                   | 配置場所                | 特徴            |
| ------------------ | ---------------------- | ----------------------- | --------------- |
| 形式バリデーション | 入力値の形式チェック   | `format_validator.go`   | DB アクセス不要 |
| 状態バリデーション | DB の状態を使った検証  | `state_validator.go` または UseCase | DB アクセス必要 |

### 構造体の命名規則

- **命名規則**: `{Action}Validator`（例: `CreateValidator`, `UpdateValidator`）
- **Input/Result パターン**: 入力パラメータと結果を専用の構造体で定義

### 状態バリデーションの配置場所

状態バリデーションは `state_validator.go` または UseCase のどちらかに配置します。

**判断基準**: **「検証失敗時に DB を更新する必要があるか？」**

| 検証失敗時の DB 更新 | 配置場所             | 理由                                               |
| -------------------- | -------------------- | -------------------------------------------------- |
| 不要                 | `state_validator.go` | UseCase をシンプルに保つため                       |
| 必要                 | UseCase              | トランザクション内で検証と更新を行う必要があるため |

**state_validator.go で行うべき検証**:

| 検証内容                   | 失敗時の動作     | 理由                   |
| -------------------------- | ---------------- | ---------------------- |
| ユーザー存在チェック       | エラーメッセージ | DB 更新なし            |
| メールアドレス重複チェック | エラーメッセージ | DB 更新なし            |
| メール確認完了チェック     | エラーメッセージ | DB 更新なし            |
| コード一致チェック         | エラーメッセージ | DB 更新なし（※注参照） |
| パスワード照合             | エラーメッセージ | DB 更新なし            |

※注: コード検証で「試行回数インクリメント」が必要な場合は UseCase で行う

**UseCase で行うべき検証**:

| 検証内容                       | 失敗時の動作           | 理由                   |
| ------------------------------ | ---------------------- | ---------------------- |
| ログインコード検証             | 試行回数インクリメント | 失敗時に DB 更新が必要 |
| 新規登録確認コード検証         | 試行回数インクリメント | 失敗時に DB 更新が必要 |
| パスワードリセットトークン検証 | トークン使用済みマーク | 失敗時に DB 更新が必要 |

### エラー表示方法の使い分け

| エラー種類           | 表示方法            | 使い分け                                           |
| -------------------- | ------------------- | -------------------------------------------------- |
| **フィールドエラー** | `FormErrors.Fields` | 特定の入力フィールドに関連するエラー               |
| **グローバルエラー** | `FormErrors.Global` | フォーム全体に関連するエラー（同じページに留まる） |
| **Flash メッセージ** | `session.Flash`     | リダイレクト後に表示するメッセージ（成功/エラー）  |
| **ログのみ**         | `slog.Error`        | 開発者向け情報（ユーザーには一般メッセージを表示） |

**判断フローチャート**:

```
フォームを再表示する？
├─ Yes → FormErrors（Fields または Global）
│    └─ 特定フィールドに関連？ → AddFieldError（例: ユーザー名重複）
│    └─ フォーム全体に関連？  → AddGlobalError（例: 確認コード不一致）
└─ No（リダイレクトする）→ Flash
     └─ 成功 → FlashSuccess
     └─ エラー → FlashError
```

### メッセージの国際化

バリデーションメッセージは必ず `i18n.T(ctx, "message_id")` を使用します。

## 実装例

### 形式バリデーション（format_validator.go）

DB を使った検証が不要な場合は、形式バリデーションのみを実装します。

```go
// internal/handler/password_reset/format_validator.go
package password_reset

import (
    "context"
    "regexp"

    "github.com/annict/annict/go/internal/i18n"
    "github.com/annict/annict/go/internal/session"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// CreateFormatValidator はパスワードリセット申請の形式バリデーションを行う
type CreateFormatValidator struct{}

// NewCreateFormatValidator は CreateFormatValidator を生成する
func NewCreateFormatValidator() *CreateFormatValidator {
    return &CreateFormatValidator{}
}

// CreateFormatValidatorInput はバリデーションの入力パラメータ
type CreateFormatValidatorInput struct {
    Email string
}

// CreateFormatValidatorResult はバリデーションの結果
type CreateFormatValidatorResult struct {
    FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateFormatValidator) Validate(ctx context.Context, input CreateFormatValidatorInput) *CreateFormatValidatorResult {
    formErrors := session.NewFormErrors()

    // 必須チェック
    if input.Email == "" {
        formErrors.AddFieldError("email", i18n.T(ctx, "password_reset_email_required"))
        return &CreateFormatValidatorResult{FormErrors: formErrors}
    }

    // フォーマットチェック
    if !emailRegex.MatchString(input.Email) {
        formErrors.AddFieldError("email", i18n.T(ctx, "password_reset_email_invalid"))
    }

    // 文字数制限
    if len(input.Email) > 255 {
        formErrors.AddFieldError("email", i18n.T(ctx, "password_reset_email_too_long"))
    }

    return &CreateFormatValidatorResult{FormErrors: formErrors}
}
```

### 状態バリデーションを含むバリデーター（state_validator.go）

DB を使った検証が必要な場合は、状態バリデーションを実装します。

```go
// internal/handler/sign_in/state_validator.go
package sign_in

import (
    "context"
    "errors"

    "github.com/annict/annict/go/internal/auth"
    "github.com/annict/annict/go/internal/i18n"
    "github.com/annict/annict/go/internal/model"
    "github.com/annict/annict/go/internal/repository"
    "github.com/annict/annict/go/internal/session"
)

// バリデーションのエラー定義
var (
    ErrUserNotFound    = errors.New("ユーザーが見つかりません")
    ErrInvalidPassword = errors.New("パスワードが正しくありません")
)

// CreateStateValidator はサインインの状態バリデーションを行う
type CreateStateValidator struct {
    userRepo *repository.UserRepository
}

// NewCreateStateValidator は CreateStateValidator を生成する
func NewCreateStateValidator(userRepo *repository.UserRepository) *CreateStateValidator {
    return &CreateStateValidator{
        userRepo: userRepo,
    }
}

// CreateStateValidatorInput はバリデーションの入力パラメータ
type CreateStateValidatorInput struct {
    EmailOrUsername string
    Password       string
}

// CreateStateValidatorResult はバリデーションの結果
type CreateStateValidatorResult struct {
    User       *model.User
    FormErrors *session.FormErrors
    Err        error
}

// Validate はバリデーションを行う
func (v *CreateStateValidator) Validate(ctx context.Context, input CreateStateValidatorInput) *CreateStateValidatorResult {
    formErrors := session.NewFormErrors()

    // ユーザー検索
    user, err := v.userRepo.GetByEmailOrUsername(ctx, input.EmailOrUsername)
    if err != nil {
        if err == repository.ErrNotFound {
            formErrors.AddGlobalError(i18n.T(ctx, "sign_in_error_invalid_credentials"))
            return &CreateStateValidatorResult{FormErrors: formErrors, Err: ErrUserNotFound}
        }
        return &CreateStateValidatorResult{Err: err}
    }

    // パスワード検証
    if err := auth.CheckPassword(user.EncryptedPassword, input.Password); err != nil {
        formErrors.AddGlobalError(i18n.T(ctx, "sign_in_error_invalid_credentials"))
        return &CreateStateValidatorResult{FormErrors: formErrors, Err: ErrInvalidPassword}
    }

    return &CreateStateValidatorResult{User: user}
}
```

### 複数フィールドの形式バリデーション

```go
// internal/handler/password/format_validator.go
package password

import (
    "context"

    "github.com/annict/annict/go/internal/i18n"
    "github.com/annict/annict/go/internal/session"
)

// UpdateFormatValidator はパスワード更新の形式バリデーションを行う
type UpdateFormatValidator struct{}

// NewUpdateFormatValidator は UpdateFormatValidator を生成する
func NewUpdateFormatValidator() *UpdateFormatValidator {
    return &UpdateFormatValidator{}
}

// UpdateFormatValidatorInput はバリデーションの入力パラメータ
type UpdateFormatValidatorInput struct {
    Password             string
    PasswordConfirmation string
}

// UpdateFormatValidatorResult はバリデーションの結果
type UpdateFormatValidatorResult struct {
    FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *UpdateFormatValidator) Validate(ctx context.Context, input UpdateFormatValidatorInput) *UpdateFormatValidatorResult {
    formErrors := session.NewFormErrors()

    // 必須チェック
    if input.Password == "" {
        formErrors.AddFieldError("password", i18n.T(ctx, "update_password_password_required"))
    }

    if input.PasswordConfirmation == "" {
        formErrors.AddFieldError("password_confirmation", i18n.T(ctx, "update_password_password_confirmation_required"))
    }

    // 文字数チェック
    if len(input.Password) > 0 && len(input.Password) < 8 {
        formErrors.AddFieldError("password", i18n.T(ctx, "update_password_password_too_short"))
    }

    // パスワード一致チェック
    if input.Password != "" && input.PasswordConfirmation != "" && input.Password != input.PasswordConfirmation {
        formErrors.AddFieldError("password_confirmation", i18n.T(ctx, "update_password_password_mismatch"))
    }

    return &UpdateFormatValidatorResult{FormErrors: formErrors}
}
```

## ハンドラーでの使用

### 基本パターン

```go
// internal/handler/sign_in/create.go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 入力データを作成
    input := CreateFormatValidatorInput{
        EmailOrUsername: r.FormValue("email_username"),
        Password:        r.FormValue("password"),
    }

    // 1. 形式バリデーション
    formatResult := h.formatValidator.Validate(ctx, input)
    if formatResult.FormErrors != nil && formatResult.FormErrors.HasErrors() {
        h.renderForm(w, ctx, csrfToken, input.EmailOrUsername, formatResult.FormErrors)
        return
    }

    // 2. 状態バリデーション
    stateInput := CreateStateValidatorInput{
        EmailOrUsername: input.EmailOrUsername,
        Password:        input.Password,
    }
    stateResult := h.stateValidator.Validate(ctx, stateInput)
    if stateResult.FormErrors != nil && stateResult.FormErrors.HasErrors() {
        h.renderForm(w, ctx, csrfToken, input.EmailOrUsername, stateResult.FormErrors)
        return
    }
    if stateResult.Err != nil {
        // システムエラー
        slog.ErrorContext(ctx, "バリデーションでエラーが発生", "error", stateResult.Err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // 認証成功後の処理（UseCase）
    // ...
}
```

### Handler の依存性

```go
// internal/handler/sign_in/handler.go
type Handler struct {
    cfg              *config.Config
    sessionMgr       *session.Manager
    formatValidator  *CreateFormatValidator  // 形式バリデーター
    stateValidator   *CreateStateValidator   // 状態バリデーター
    createSessionUC  *usecase.CreateSessionUsecase
}

func NewHandler(
    cfg *config.Config,
    sessionMgr *session.Manager,
    userRepo *repository.UserRepository,
    createSessionUC *usecase.CreateSessionUsecase,
) *Handler {
    return &Handler{
        cfg:              cfg,
        sessionMgr:       sessionMgr,
        formatValidator:  NewCreateFormatValidator(),
        stateValidator:   NewCreateStateValidator(userRepo),
        createSessionUC:  createSessionUC,
    }
}
```

### PATCHメソッドでの使用（Method Override）

```go
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Method Overrideミドルウェアにより、POSTリクエストがPATCHに変換される
    // フォーム: <input type="hidden" name="_method" value="PATCH" />

    input := UpdateFormatValidatorInput{
        Password:             r.FormValue("password"),
        PasswordConfirmation: r.FormValue("password_confirmation"),
    }

    // 形式バリデーション
    result := h.formatValidator.Validate(ctx, input)
    if result.FormErrors != nil && result.FormErrors.HasErrors() {
        h.renderForm(w, ctx, csrfToken, result.FormErrors)
        return
    }

    // ビジネスロジック（UseCase）
    // ...
}
```

## テスト

### テスト方針

バリデーションのテストは、形式バリデーションと状態バリデーションに分けて実装します。

| テスト対象           | ファイル                    | 特徴                                   |
| -------------------- | --------------------------- | -------------------------------------- |
| 形式バリデーション   | `format_validator_test.go`  | DB アクセス不要、高速に実行可能        |
| 状態バリデーション   | `state_validator_test.go`   | DB アクセス必要、テスト用 DB を使用    |
| ハンドラーの振る舞い | `create_test.go` 等         | E2E テスト、正常系・代表的な異常系のみ |

**理由**:

- **問題の特定**: テスト失敗時にどの検証の問題か即座に分かる
- **保守性向上**: テストファイルの管理が容易
- **実行速度**: 形式バリデーションのテストは DB 不要で高速に実行可能

### 形式バリデーションのテスト

```go
// internal/handler/sign_in/format_validator_test.go
func TestCreateFormatValidator_Validate(t *testing.T) {
    tests := []struct {
        name          string
        input         CreateFormatValidatorInput
        wantErrors    bool
        expectedField string
    }{
        {
            name: "有効な入力",
            input: CreateFormatValidatorInput{
                EmailOrUsername: "user@example.com",
                Password:        "password123",
            },
            wantErrors: false,
        },
        {
            name: "メールアドレスが空",
            input: CreateFormatValidatorInput{
                EmailOrUsername: "",
                Password:        "password123",
            },
            wantErrors:    true,
            expectedField: "email_username",
        },
        {
            name: "パスワードが空",
            input: CreateFormatValidatorInput{
                EmailOrUsername: "user@example.com",
                Password:        "",
            },
            wantErrors:    true,
            expectedField: "password",
        },
        {
            name: "両方とも空",
            input: CreateFormatValidatorInput{
                EmailOrUsername: "",
                Password:        "",
            },
            wantErrors: true,
        },
    }

    // DBアクセス不要でテスト
    validator := NewCreateFormatValidator()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            ctx = i18n.WithLocale(ctx, "ja")

            result := validator.Validate(ctx, tt.input)

            if tt.wantErrors {
                if result.FormErrors == nil || !result.FormErrors.HasErrors() {
                    t.Error("expected errors, got none")
                }
                if tt.expectedField != "" && !result.FormErrors.HasFieldError(tt.expectedField) {
                    t.Errorf("expected field error for %q", tt.expectedField)
                }
            } else {
                if result.FormErrors != nil && result.FormErrors.HasErrors() {
                    t.Errorf("expected no errors, got %v", result.FormErrors)
                }
            }
        })
    }
}
```

### 状態バリデーションのテスト（DB必要）

```go
// internal/handler/sign_in/state_validator_test.go
func TestCreateStateValidator_Validate(t *testing.T) {
    // テストDBとトランザクションをセットアップ
    db, tx := testutil.SetupTestDB(t)

    // テストユーザーを作成
    testutil.NewUserBuilder(t, tx).
        WithEmail("test@example.com").
        WithPassword("password123").
        Build()

    userRepo := repository.NewUserRepository(db).WithTx(tx)
    validator := NewCreateStateValidator(userRepo)

    t.Run("有効な認証情報", func(t *testing.T) {
        ctx := context.Background()
        input := CreateStateValidatorInput{
            EmailOrUsername: "test@example.com",
            Password:        "password123",
        }

        result := validator.Validate(ctx, input)

        if result.FormErrors != nil && result.FormErrors.HasErrors() {
            t.Errorf("unexpected form errors: %v", result.FormErrors)
        }
        if result.User == nil {
            t.Error("expected user, got nil")
        }
    })

    t.Run("無効なパスワード", func(t *testing.T) {
        ctx := context.Background()
        input := CreateStateValidatorInput{
            EmailOrUsername: "test@example.com",
            Password:        "wrongpassword",
        }

        result := validator.Validate(ctx, input)

        if result.User != nil {
            t.Error("expected nil user")
        }
        if result.FormErrors == nil || !result.FormErrors.HasErrors() {
            t.Error("expected form errors")
        }
    })
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

### 1. 形式バリデーションと状態バリデーションを分離

```go
// ✅ Good: 形式バリデーションと状態バリデーションを分離
// format_validator.go: 形式チェックのみ（DB不要）
type CreateFormatValidator struct{}

func (v *CreateFormatValidator) Validate(ctx context.Context, input CreateFormatValidatorInput) *CreateFormatValidatorResult {
    // 必須チェック、フォーマット検証のみ
}

// state_validator.go: 状態チェック（DB必要）
type CreateStateValidator struct {
    userRepo *repository.UserRepository
}

func (v *CreateStateValidator) Validate(ctx context.Context, input CreateStateValidatorInput) *CreateStateValidatorResult {
    // ユーザー存在チェック、パスワード照合など
}

// ハンドラーで順次実行
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. 形式バリデーション
    formatResult := h.formatValidator.Validate(ctx, formatInput)
    if formatResult.FormErrors.HasErrors() { ... }

    // 2. 状態バリデーション
    stateResult := h.stateValidator.Validate(ctx, stateInput)
    if stateResult.FormErrors.HasErrors() { ... }
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

func (v *CreateFormatValidator) Validate(ctx context.Context, input CreateFormatValidatorInput) *CreateFormatValidatorResult {
    // emailRegexを使用
}
```

### 5. Result 構造体でバリデーション結果を返す

```go
// ✅ Good: 結果を構造体で返す
type CreateStateValidatorResult struct {
    User       *model.User        // 成功時のデータ
    FormErrors *session.FormErrors // フォームエラー
    Err        error               // システムエラー
}

func (v *CreateStateValidator) Validate(ctx context.Context, input CreateStateValidatorInput) *CreateStateValidatorResult {
    // ...
    return &CreateStateValidatorResult{User: user}
}
```

## 利点

1. **関心の分離**: 形式バリデーションと状態バリデーションが明確に分離される
2. **テストしやすい**: 形式バリデーションは DB モックなしでテストできる
3. **依存が明確**: バリデーターの依存関係が一目でわかる
4. **ハンドラーの見通しが良くなる**: ハンドラーはHTTP処理とビジネスロジックに専念
5. **再利用可能**: 同じバリデーターを複数のハンドラーで使用可能
