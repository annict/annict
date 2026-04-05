# バリデーションガイド

このドキュメントは、Go 版 Annict でのバリデーションのベストプラクティスを説明します。

## 概要

フォームからの入力値の検証は、`internal/validator/` パッケージ（Application 層）に配置し、UseCase から呼び出します。形式バリデーション（入力値の形式チェック）と状態バリデーション（DB を使った検証）を同じファイルに実装します。

### ファイル構成

```
internal/validator/
├── sign_in.go               # サインインのバリデーション
├── sign_in_test.go          # サインインのバリデーションテスト
├── sign_in_password.go      # パスワードサインインのバリデーション
├── sign_in_password_test.go # パスワードサインインのバリデーションテスト
├── password_reset.go        # パスワードリセットのバリデーション
├── password_reset_test.go   # パスワードリセットのバリデーションテスト
├── password.go              # パスワード更新のバリデーション
└── password_test.go         # パスワード更新のバリデーションテスト
```

### バリデーションの分類

バリデーションは以下の 2 種類に分類されますが、同じファイルに実装します：

| 種類               | 責務                  | 特徴            |
| ------------------ | --------------------- | --------------- |
| 形式バリデーション | 入力値の形式チェック  | DB アクセス不要 |
| 状態バリデーション | DB の状態を使った検証 | DB アクセス必要 |

### 命名規則

- **配置**: `internal/validator/`（Application 層）
- **ファイル名**: リソース名と対応（例: `sign_in.go`, `password_reset.go`）
- **構造体名**: `{Action}{Resource}Validator`（例: `CreateSignInValidator`, `UpdatePasswordValidator`）
- **入力構造体**: `{Action}{Resource}ValidatorInput`
- **結果構造体**: `{Action}{Resource}ValidatorResult`
- **呼び出し元**: UseCase（Handler から直接呼び出さない）

### 状態バリデーションの配置場所

状態バリデーションは `validator` パッケージまたは UseCase のどちらかに配置します。

**判断基準**: **「検証失敗時に DB を更新する必要があるか？」**

| 検証失敗時の DB 更新 | 配置場所  | 理由                                               |
| -------------------- | --------- | -------------------------------------------------- |
| 不要                 | validator | UseCase をシンプルに保つため                       |
| 必要                 | UseCase   | トランザクション内で検証と更新を行う必要があるため |

**validator で行うべき検証**:

| 検証内容                   | 失敗時の動作     | 理由                   |
| -------------------------- | ---------------- | ---------------------- |
| ユーザー存在チェック       | エラーメッセージ | DB 更新なし            |
| メールアドレス重複チェック | エラーメッセージ | DB 更新なし            |
| アットネーム重複チェック   | エラーメッセージ | DB 更新なし            |
| メール確認完了チェック     | エラーメッセージ | DB 更新なし            |
| コード一致チェック         | エラーメッセージ | DB 更新なし（※注参照） |
| パスワード照合             | エラーメッセージ | DB 更新なし            |

※注: コード検証で「試行回数インクリメント」が必要な場合は UseCase で行う

**UseCase で行うべき検証**:

| 検証内容           | 失敗時の動作           | 理由                   |
| ------------------ | ---------------------- | ---------------------- |
| ログインコード検証 | 試行回数インクリメント | 失敗時に DB 更新が必要 |

※注: 「リカバリーコード消費」や「トークン使用済みマーク」は検証成功後の処理であり、バリデーションではなく UseCase の永続化処理として扱う。検証自体は validator で行い、成功後に UseCase を呼び出す。

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

### 基本的なバリデーター（形式バリデーションのみ）

```go
// internal/validator/password_reset.go
package validator

import (
    "context"
    "strings"

    "github.com/annict/annict/go/internal/i18n"
    "github.com/annict/annict/go/internal/session"
)

// CreatePasswordResetValidator はパスワードリセット申請フォームのバリデーションを行う
type CreatePasswordResetValidator struct{}

// NewCreatePasswordResetValidator は CreatePasswordResetValidator を生成する
func NewCreatePasswordResetValidator() *CreatePasswordResetValidator {
    return &CreatePasswordResetValidator{}
}

// CreatePasswordResetValidatorInput はバリデーションの入力パラメータ
type CreatePasswordResetValidatorInput struct {
    Email string
}

// CreatePasswordResetValidatorResult はバリデーションの結果
type CreatePasswordResetValidatorResult struct {
    FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreatePasswordResetValidator) Validate(ctx context.Context, input CreatePasswordResetValidatorInput) *CreatePasswordResetValidatorResult {
    formErrors := &session.FormErrors{}

    if strings.TrimSpace(input.Email) == "" {
        formErrors.AddFieldError("email", i18n.T(ctx, "password_reset_email_required"))
    }

    if formErrors.HasErrors() {
        return &CreatePasswordResetValidatorResult{FormErrors: formErrors}
    }

    return &CreatePasswordResetValidatorResult{}
}
```

### 状態バリデーションを含むバリデーター

DB を使った検証が必要な場合は、Repository を依存として注入します。

```go
// internal/validator/sign_in_password.go
package validator

import (
    "context"
    "errors"
    "strings"

    "github.com/annict/annict/go/internal/auth"
    "github.com/annict/annict/go/internal/i18n"
    "github.com/annict/annict/go/internal/model"
    "github.com/annict/annict/go/internal/repository"
    "github.com/annict/annict/go/internal/session"
)

var (
    ErrUserNotFound    = errors.New("ユーザーが見つかりません")
    ErrInvalidPassword = errors.New("パスワードが正しくありません")
)

// CreateSignInPasswordValidator はパスワードサインインのバリデーションを行う
type CreateSignInPasswordValidator struct {
    userRepo *repository.UserRepository
}

// NewCreateSignInPasswordValidator は CreateSignInPasswordValidator を生成する
func NewCreateSignInPasswordValidator(userRepo *repository.UserRepository) *CreateSignInPasswordValidator {
    return &CreateSignInPasswordValidator{
        userRepo: userRepo,
    }
}

// CreateSignInPasswordValidatorInput はバリデーションの入力パラメータ
type CreateSignInPasswordValidatorInput struct {
    EmailOrUsername string
    Password       string
}

// CreateSignInPasswordValidatorResult はバリデーションの結果
type CreateSignInPasswordValidatorResult struct {
    User       *model.User
    FormErrors *session.FormErrors
    Err        error
}

// Validate はバリデーションを行う
func (v *CreateSignInPasswordValidator) Validate(ctx context.Context, input CreateSignInPasswordValidatorInput) *CreateSignInPasswordValidatorResult {
    // 1. 形式バリデーション
    formErrors := &session.FormErrors{}

    if strings.TrimSpace(input.EmailOrUsername) == "" {
        formErrors.AddFieldError("email_or_username", i18n.T(ctx, "error_required"))
    }

    if input.Password == "" {
        formErrors.AddFieldError("password", i18n.T(ctx, "error_required"))
    }

    if formErrors.HasErrors() {
        return &CreateSignInPasswordValidatorResult{FormErrors: formErrors}
    }

    // 2. 状態バリデーション（DB検証）
    user, err := v.userRepo.GetByEmailOrUsername(ctx, input.EmailOrUsername)
    if err != nil {
        formErrors.AddGlobalError(i18n.T(ctx, "error_invalid_credentials"))
        return &CreateSignInPasswordValidatorResult{FormErrors: formErrors, Err: ErrUserNotFound}
    }

    // パスワード検証
    if err := auth.CheckPassword(user.EncryptedPassword, input.Password); err != nil {
        formErrors.AddGlobalError(i18n.T(ctx, "error_invalid_credentials"))
        return &CreateSignInPasswordValidatorResult{FormErrors: formErrors, Err: ErrInvalidPassword}
    }

    return &CreateSignInPasswordValidatorResult{User: user}
}
```

### 複数フィールドのバリデーター

```go
// internal/validator/password.go
package validator

import (
    "context"
    "strings"

    "github.com/annict/annict/go/internal/i18n"
    "github.com/annict/annict/go/internal/session"
)

// UpdatePasswordValidator はパスワード更新のバリデーションを行う
type UpdatePasswordValidator struct{}

// NewUpdatePasswordValidator は UpdatePasswordValidator を生成する
func NewUpdatePasswordValidator() *UpdatePasswordValidator {
    return &UpdatePasswordValidator{}
}

// UpdatePasswordValidatorInput はバリデーションの入力パラメータ
type UpdatePasswordValidatorInput struct {
    Password             string
    PasswordConfirmation string
}

// UpdatePasswordValidatorResult はバリデーションの結果
type UpdatePasswordValidatorResult struct {
    FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *UpdatePasswordValidator) Validate(ctx context.Context, input UpdatePasswordValidatorInput) *UpdatePasswordValidatorResult {
    formErrors := &session.FormErrors{}

    if strings.TrimSpace(input.Password) == "" {
        formErrors.AddFieldError("password", i18n.T(ctx, "error_required"))
    }

    if strings.TrimSpace(input.PasswordConfirmation) == "" {
        formErrors.AddFieldError("password_confirmation", i18n.T(ctx, "error_required"))
    }

    if len(input.Password) > 0 && len(input.Password) < 8 {
        formErrors.AddFieldError("password", i18n.T(ctx, "error_password_too_short"))
    }

    if input.Password != "" && input.PasswordConfirmation != "" && input.Password != input.PasswordConfirmation {
        formErrors.AddFieldError("password_confirmation", i18n.T(ctx, "error_password_mismatch"))
    }

    return &UpdatePasswordValidatorResult{FormErrors: formErrors}
}
```

## UseCase からの呼び出し

Validator は UseCase から呼び出します。Handler から直接呼び出しません。

### UseCase での使用パターン

```go
// internal/usecase/create_password_reset_token.go
package usecase

import (
    "context"
    "database/sql"

    "github.com/annict/annict/go/internal/repository"
    "github.com/annict/annict/go/internal/session"
    "github.com/annict/annict/go/internal/validator"
)

type CreatePasswordResetTokenUsecase struct {
    db        *sql.DB
    userRepo  *repository.UserRepository
    validator *validator.CreatePasswordResetValidator
    // ...
}

type CreatePasswordResetTokenInput struct {
    Email string
}

type CreatePasswordResetTokenResult struct {
    FormErrors *session.FormErrors
}

func (uc *CreatePasswordResetTokenUsecase) Execute(ctx context.Context, input CreatePasswordResetTokenInput) (*CreatePasswordResetTokenResult, error) {
    // 1. バリデーション（Validator を呼び出し）
    valResult := uc.validator.Validate(ctx, validator.CreatePasswordResetValidatorInput{
        Email: input.Email,
    })
    if valResult.FormErrors != nil && valResult.FormErrors.HasErrors() {
        return &CreatePasswordResetTokenResult{FormErrors: valResult.FormErrors}, nil
    }

    // 2. ビジネスロジック + 永続化
    // ...

    return &CreatePasswordResetTokenResult{}, nil
}
```

### Handler での使用パターン

Handler は UseCase を呼び出し、結果のバリデーションエラーに基づいてレスポンスを返します。

```go
// internal/handler/password_reset/create.go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // UseCaseを呼び出し（バリデーション + ビジネスロジック）
    result, err := h.createPasswordResetTokenUC.Execute(ctx, usecase.CreatePasswordResetTokenInput{
        Email: r.FormValue("email"),
    })
    if err != nil {
        // システムエラー
        slog.ErrorContext(ctx, "パスワードリセット処理でエラーが発生", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    if result.FormErrors != nil && result.FormErrors.HasErrors() {
        // バリデーションエラー → フォーム再表示
        // ...
        return
    }

    // 成功 → リダイレクト
    http.Redirect(w, r, "/password/reset/sent", http.StatusSeeOther)
}
```

## テスト

### テスト方針

バリデーションのテストは `internal/validator/` パッケージ内の `*_test.go` ファイルに実装します。

| テスト対象                    | ファイル                       | 特徴                                 |
| ----------------------------- | ------------------------------ | ------------------------------------ |
| バリデーション                | `validator/*_test.go`          | 形式・状態バリデーションを統合テスト |
| UseCase（バリデーション含む） | `usecase/*_test.go`            | UseCase 経由のバリデーションテスト   |
| Handler の振る舞い            | `handler/{resource}/*_test.go` | E2E テスト、正常系・代表的な異常系   |

### バリデーションのテスト

```go
// internal/validator/sign_in_test.go
func TestCreateSignInValidator_Validate(t *testing.T) {
    t.Parallel()

    validator := NewCreateSignInValidator()

    tests := []struct {
        name       string
        input      CreateSignInValidatorInput
        wantErrors bool
        wantField  string
    }{
        {
            name:       "有効な入力",
            input:      CreateSignInValidatorInput{Email: "user@example.com"},
            wantErrors: false,
        },
        {
            name:       "メールアドレスが空",
            input:      CreateSignInValidatorInput{Email: ""},
            wantErrors: true,
            wantField:  "email",
        },
        {
            name:       "スペースのみ",
            input:      CreateSignInValidatorInput{Email: "   "},
            wantErrors: true,
            wantField:  "email",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            result := validator.Validate(ctx, tt.input)

            if tt.wantErrors {
                if result.FormErrors == nil || !result.FormErrors.HasErrors() {
                    t.Error("expected errors, got none")
                }
                if tt.wantField != "" && !result.FormErrors.HasFieldError(tt.wantField) {
                    t.Errorf("expected field error for %q", tt.wantField)
                }
            } else {
                if result.FormErrors != nil && result.FormErrors.HasErrors() {
                    t.Errorf("unexpected errors: %v", result.FormErrors)
                }
            }
        })
    }
}
```

### 状態バリデーションのテスト（DB 必要）

```go
// internal/validator/sign_in_password_test.go
func TestCreateSignInPasswordValidator_Validate(t *testing.T) {
    t.Parallel()

    db, tx := testutil.SetupTestDB(t)

    // テストユーザーを作成
    testutil.NewUserBuilder(t, tx).
        WithEmail("test@example.com").
        WithPassword("password123").
        Build()

    userRepo := repository.NewUserRepository(db).WithTx(tx)
    validator := NewCreateSignInPasswordValidator(userRepo)

    t.Run("有効な認証情報", func(t *testing.T) {
        ctx := context.Background()
        result := validator.Validate(ctx, CreateSignInPasswordValidatorInput{
            EmailOrUsername: "test@example.com",
            Password:       "password123",
        })

        if result.FormErrors != nil && result.FormErrors.HasErrors() {
            t.Errorf("unexpected form errors: %v", result.FormErrors)
        }
        if result.User == nil {
            t.Error("expected user, got nil")
        }
    })

    t.Run("無効なパスワード", func(t *testing.T) {
        ctx := context.Background()
        result := validator.Validate(ctx, CreateSignInPasswordValidatorInput{
            EmailOrUsername: "test@example.com",
            Password:       "wrongpassword",
        })

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

### 1. Validator は Application 層に配置し、UseCase から呼び出す

```go
// ✅ Good: internal/validator/ に配置し、UseCase から呼び出す
// internal/validator/sign_in.go
package validator

type CreateSignInValidator struct{}

// internal/usecase/send_sign_in_code.go
type SendSignInCodeUsecase struct {
    validator *validator.CreateSignInValidator
}

func (uc *SendSignInCodeUsecase) Execute(ctx context.Context, input SendSignInCodeInput) (*SendSignInCodeResult, error) {
    valResult := uc.validator.Validate(ctx, validator.CreateSignInValidatorInput{
        Email: input.Email,
    })
    // ...
}
```

```go
// ❌ Bad: Handler から直接 Validator を呼び出す
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    result := h.validator.Validate(ctx, input) // Handler が直接呼び出している
    // ...
}
```

### 2. 国際化を徹底

```go
// ❌ Bad: ハードコードされたメッセージ
formErrors.AddFieldError("email", "メールアドレスを入力してください")

// ✅ Good: 国際化された翻訳
formErrors.AddFieldError("email", i18n.T(ctx, "error_required"))
```

### 3. 早期リターンでネストを減らす

```go
// ❌ Bad: ネストが深い
func (v *CreatePasswordResetValidator) Validate(ctx context.Context, input CreatePasswordResetValidatorInput) *CreatePasswordResetValidatorResult {
    formErrors := &session.FormErrors{}
    if input.Email != "" {
        if emailRegex.MatchString(input.Email) {
            // OK
        } else {
            formErrors.AddFieldError("email", "...")
        }
    } else {
        formErrors.AddFieldError("email", "...")
    }
    return &CreatePasswordResetValidatorResult{FormErrors: formErrors}
}

// ✅ Good: 早期リターンでシンプル
func (v *CreatePasswordResetValidator) Validate(ctx context.Context, input CreatePasswordResetValidatorInput) *CreatePasswordResetValidatorResult {
    formErrors := &session.FormErrors{}

    if strings.TrimSpace(input.Email) == "" {
        formErrors.AddFieldError("email", i18n.T(ctx, "error_required"))
        return &CreatePasswordResetValidatorResult{FormErrors: formErrors}
    }

    return &CreatePasswordResetValidatorResult{}
}
```

### 4. 正規表現はパッケージレベルで定義

```go
// ✅ Good: 正規表現のコンパイルを1回だけ実行
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (v *CreatePasswordResetValidator) Validate(ctx context.Context, input CreatePasswordResetValidatorInput) *CreatePasswordResetValidatorResult {
    // emailRegexを使用
}
```

### 5. Result 構造体でバリデーション結果を返す

```go
// ✅ Good: 結果を構造体で返す
type CreateSignInPasswordValidatorResult struct {
    User       *model.User        // 成功時のデータ
    FormErrors *session.FormErrors // フォームエラー
    Err        error               // システムエラー
}

func (v *CreateSignInPasswordValidator) Validate(ctx context.Context, input CreateSignInPasswordValidatorInput) *CreateSignInPasswordValidatorResult {
    // ...
    return &CreateSignInPasswordValidatorResult{User: user}
}
```

## 利点

1. **関心の分離**: バリデーションが Application 層に配置され、UseCase から統括的に呼び出される
2. **再利用可能**: 同じバリデーターを複数の UseCase から利用できる
3. **テストしやすい**: Validator を独立してテストでき、UseCase テストではバリデーション込みの統合テストが可能
4. **依存が明確**: バリデーターの依存関係が一目でわかる
5. **Handler がシンプル**: Handler は HTTP の入出力変換のみに専念できる
