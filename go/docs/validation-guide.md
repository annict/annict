# バリデーションガイド

このドキュメントは、Go 版 Annict でのバリデーションのベストプラクティスを説明します。

## 概要

フォームからの入力値の検証は、`internal/validator/` パッケージの**バリデーター**で実装します。形式バリデーション（入力値の形式チェック）と状態バリデーション（DB を使った検証）を同じパッケージに配置することで、「どこに書くべきか」の判断コストをゼロにします。

### バリデーターの配置先

すべてのバリデーターは `internal/validator/` パッケージに配置します。形式バリデーションのみの場合も、状態バリデーションを含む場合も同じパッケージです。これにより：

- **判断コストがゼロ**: バリデーションは常に `internal/validator/` に配置する
- **進化に強い**: 形式バリデーションに後から DB チェックが追加されても、ファイル移動が不要
- **一箇所で把握できる**: バリデーションルールを確認したいとき `internal/validator/` だけ見ればよい
- **アーキテクチャの強制**: Handler パッケージから `repository`・`validator` パッケージの import を完全に排除し、depguard で強制できる

### ファイル構成

```
internal/validator/
├── sign_in_password.go        # バリデーション（形式チェック + DBを使った検証）
├── sign_in_password_test.go   # バリデーションのテスト
├── password_reset.go          # バリデーション（形式チェックのみ）
├── password_reset_test.go     # バリデーションのテスト
├── password.go                # パスワード更新のバリデーション
└── password_test.go           # パスワード更新のバリデーションのテスト

internal/usecase/
├── authenticate_by_password.go       # UseCase（Validatorを外部から受け取り、呼び出す）
└── authenticate_by_password_test.go

internal/handler/sign_in_password/
├── handler.go            # Handler構造体と依存性（UseCaseを外部から受け取る）
├── new.go                # フォーム表示
└── create.go             # 作成処理
```

### バリデーションの分類

バリデーションは以下の 2 種類に分類されますが、同じファイルに実装します：

| 種類               | 責務                  | 特徴            |
| ------------------ | --------------------- | --------------- |
| 形式バリデーション | 入力値の形式チェック  | DB アクセス不要 |
| 状態バリデーション | DB の状態を使った検証 | DB アクセス必要 |

### 構造体の命名規則

- **命名規則**: `{Resource}{Action}Validator`（例: `SignInPasswordCreateValidator`, `PasswordResetCreateValidator`, `PasswordUpdateValidator`）
- **コンストラクタ**: `New{Resource}{Action}Validator`（例: `NewSignInPasswordCreateValidator`）
- **入力構造体**: `{Resource}{Action}ValidatorInput`
- **出力構造体**: `{Resource}{Action}ValidateOutput`（データ取得を伴うバリデーターのみ）
- **戻り値**: Go の慣習に従った `(data, error)` の 2 値返し。データを返す必要がない場合は `error` のみ
- **呼び出し元**: UseCase から呼び出される（Handler からの直接呼び出しは禁止）
- **1 つの構造体で両方のバリデーションを担当**: 形式バリデーションと状態バリデーションを `Validate` メソッド内で順次実行

### 状態バリデーションの配置場所

状態バリデーションは `internal/validator/` パッケージ内のバリデーターまたは UseCase のどちらかに配置します。

**判断基準**: **「検証失敗時に DB を更新する必要があるか？」**

| 検証失敗時の DB 更新 | 配置場所  | 理由                                               |
| -------------------- | --------- | -------------------------------------------------- |
| 不要                 | Validator | UseCase をシンプルに保つため                       |
| 必要                 | UseCase   | トランザクション内で検証と更新を行う必要があるため |

**Validator で行うべき検証**:

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

※注: 「リカバリーコード消費」や「トークン使用済みマーク」は検証成功後の処理であり、バリデーションではなく UseCase の永続化処理として扱う。検証自体は Validator で行い、成功後に UseCase を呼び出す。

### エラー型の使い分け

Validator は `error` インターフェースを返します。エラーの具体的な型によって Handler の対応が決まります。

| エラー型                 | 生成元    | 意味                             | Handler の対応                          |
| ------------------------ | --------- | -------------------------------- | --------------------------------------- |
| `*model.ValidationError` | Validator | 入力が不正（ユーザーが修正可能） | フォーム再描画（422）                   |
| `*model.AppError`        | UseCase   | 業務レベルの既知の失敗           | エラーコードに応じた処理（403, 404 等） |
| 素の `error`             | どこでも  | 予期しないシステムエラー         | 500                                     |

**`model.ValidationError` のエラー表示方法**:

| エラー種類           | 表示方法                 | 使い分け                                           |
| -------------------- | ------------------------ | -------------------------------------------------- |
| **フィールドエラー** | `ValidationError.Fields` | 特定の入力フィールドに関連するエラー               |
| **グローバルエラー** | `ValidationError.Global` | フォーム全体に関連するエラー（同じページに留まる） |
| **Flash メッセージ** | `session.FlashManager`   | リダイレクト後に表示するメッセージ（成功/エラー）  |
| **ログのみ**         | `slog.Error`             | 開発者向け情報（ユーザーには一般メッセージを表示） |

**判断フローチャート**:

```
フォームを再表示する？
├─ Yes → ValidationError（Fields または Global）
│    └─ 特定フィールドに関連？ → AddField（例: ユーザー名重複）
│    └─ フォーム全体に関連？  → AddGlobal（例: 確認コード不一致）
└─ No（リダイレクトする）→ Flash
     └─ 成功 → FlashSuccess
     └─ エラー → FlashError
```

### メッセージの国際化

バリデーションメッセージは必ず `i18n.T(ctx, "message_id")` を使用します。

## 実装例

### データを返すバリデーター（状態バリデーションあり）

状態バリデーションの過程で取得したデータを戻り値として返します。これにより UseCase 内でデータを二重に取得する必要がなくなります。

```go
// internal/validator/sign_in_password.go
package validator

import (
    "context"
    "database/sql"
    "errors"
    "strings"

    "github.com/annict/annict/go/internal/auth"
    "github.com/annict/annict/go/internal/i18n"
    "github.com/annict/annict/go/internal/model"
    "github.com/annict/annict/go/internal/repository"
)

// SignInPasswordCreateValidator はパスワードログインのバリデーションを行う
type SignInPasswordCreateValidator struct {
    userRepo *repository.UserRepository
}

// NewSignInPasswordCreateValidator は SignInPasswordCreateValidator を生成する
func NewSignInPasswordCreateValidator(userRepo *repository.UserRepository) *SignInPasswordCreateValidator {
    return &SignInPasswordCreateValidator{
        userRepo: userRepo,
    }
}

// SignInPasswordCreateValidatorInput はバリデーションの入力パラメータ
type SignInPasswordCreateValidatorInput struct {
    EmailOrUsername string
    Password        string
}

// SignInPasswordCreateValidateOutput はバリデーション成功時の出力
type SignInPasswordCreateValidateOutput struct {
    User repository.GetUserByEmailOrUsernameRow
}

// Validate はバリデーションを行い、成功時は認証済みユーザー情報を返す
func (v *SignInPasswordCreateValidator) Validate(ctx context.Context, input SignInPasswordCreateValidatorInput) (*SignInPasswordCreateValidateOutput, error) {
    // 1. 形式バリデーション
    ve := model.NewValidationError()

    if strings.TrimSpace(input.EmailOrUsername) == "" {
        ve.AddField("email_or_username", i18n.T(ctx, "sign_in_error_email_required"))
    }

    if strings.TrimSpace(input.Password) == "" {
        ve.AddField("password", i18n.T(ctx, "sign_in_error_password_required"))
    }

    if ve.HasErrors() {
        return nil, ve
    }

    // 2. 状態バリデーション（DB検証）
    user, err := v.userRepo.GetByEmailOrUsername(ctx, input.EmailOrUsername)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            ve.AddGlobal(i18n.T(ctx, "sign_in_error_invalid_credentials"))
            return nil, ve
        }
        return nil, err
    }

    // パスワード検証
    if err := auth.CheckPassword(user.EncryptedPassword, input.Password); err != nil {
        ve.AddGlobal(i18n.T(ctx, "sign_in_error_invalid_credentials"))
        return nil, ve
    }

    return &SignInPasswordCreateValidateOutput{User: user}, nil
}
```

### データを返さないバリデーター（形式バリデーションのみ）

DB を使った検証が不要な場合でも、`internal/validator/` パッケージに配置します。データを返す必要がない場合は `error` のみを返します。

```go
// internal/validator/password_reset.go
package validator

import (
    "context"
    "strings"

    "github.com/annict/annict/go/internal/i18n"
    "github.com/annict/annict/go/internal/model"
)

// PasswordResetCreateValidator はパスワードリセット申請フォームのバリデーションを行う
type PasswordResetCreateValidator struct{}

// NewPasswordResetCreateValidator は PasswordResetCreateValidator を生成する
func NewPasswordResetCreateValidator() *PasswordResetCreateValidator {
    return &PasswordResetCreateValidator{}
}

// PasswordResetCreateValidatorInput はバリデーションの入力パラメータ
type PasswordResetCreateValidatorInput struct {
    Email string
}

// Validate はバリデーションを行う
func (v *PasswordResetCreateValidator) Validate(ctx context.Context, input PasswordResetCreateValidatorInput) error {
    ve := model.NewValidationError()

    if strings.TrimSpace(input.Email) == "" {
        ve.AddField("email", i18n.T(ctx, "password_reset_email_required"))
    }

    if ve.HasErrors() {
        return ve
    }
    return nil
}
```

### 複数フィールドのバリデーター

複数フィールドを持つフォームでは、フィールド間のクロスチェック（例: パスワードと確認用パスワードの一致）も同じ Validator で行います。

```go
// internal/validator/password.go
package validator

import (
    "context"
    "strings"

    "github.com/annict/annict/go/internal/auth"
    "github.com/annict/annict/go/internal/i18n"
    "github.com/annict/annict/go/internal/model"
)

// PasswordUpdateValidator はパスワード更新フォームのバリデーションを行う
type PasswordUpdateValidator struct{}

// NewPasswordUpdateValidator は PasswordUpdateValidator を生成する
func NewPasswordUpdateValidator() *PasswordUpdateValidator {
    return &PasswordUpdateValidator{}
}

// PasswordUpdateValidatorInput はバリデーションの入力パラメータ
type PasswordUpdateValidatorInput struct {
    Token                string
    Password             string
    PasswordConfirmation string
}

// Validate はバリデーションを行う
func (v *PasswordUpdateValidator) Validate(ctx context.Context, input PasswordUpdateValidatorInput) error {
    ve := model.NewValidationError()

    if strings.TrimSpace(input.Token) == "" {
        ve.AddField("token", i18n.T(ctx, "password_reset_token_invalid"))
    }

    if strings.TrimSpace(input.Password) == "" {
        ve.AddField("password", i18n.T(ctx, "password_reset_password_required"))
    }

    if strings.TrimSpace(input.PasswordConfirmation) == "" {
        ve.AddField("password_confirmation", i18n.T(ctx, "password_reset_password_confirmation_required"))
    }

    // クロスフィールド検証（パスワード一致チェック）
    if strings.TrimSpace(input.Password) != "" && strings.TrimSpace(input.PasswordConfirmation) != "" && input.Password != input.PasswordConfirmation {
        ve.AddField("password_confirmation", i18n.T(ctx, "password_reset_password_mismatch"))
    }

    if ve.HasErrors() {
        return ve
    }

    // パスワード強度チェック
    if err := auth.ValidatePasswordStrength(ctx, input.Password); err != nil {
        ve.AddField("password", err.Error())
        return ve
    }

    return nil
}
```

## UseCase での使用

Validator は UseCase から呼び出されます。Handler は UseCase を呼び出すだけで、Validator に直接依存しません。

### UseCase でのバリデーション呼び出し

Validator がデータを返す場合、UseCase はその戻り値を使って以降の処理を進めます。

```go
// internal/usecase/authenticate_by_password.go
func (uc *AuthenticateByPasswordUsecase) Execute(ctx context.Context, input AuthenticateByPasswordInput) (*AuthenticateByPasswordOutput, error) {
    // 1. バリデーション（形式チェック + 存在確認 + パスワード照合）
    valOutput, err := uc.validator.Validate(ctx, validator.SignInPasswordCreateValidatorInput{
        EmailOrUsername: input.Email,
        Password:        input.Password,
    })
    if err != nil {
        return nil, err // *model.ValidationError か素の error がそのまま上がる
    }
    user := valOutput.User

    // 2. セッション作成
    sessionResult, err := uc.createSessionUC.Execute(ctx, nil, user.ID, user.EncryptedPassword)
    if err != nil {
        return nil, fmt.Errorf("セッションの作成に失敗: %w", err)
    }

    return &AuthenticateByPasswordOutput{
        PublicID: sessionResult.PublicID,
        UserID:   user.ID,
        Username: user.Username,
    }, nil
}
```

Validator がデータを返さない場合は、`error` のみを判定します。

```go
// internal/usecase/create_password_reset_token.go
func (uc *CreatePasswordResetTokenUsecase) Execute(ctx context.Context, input CreatePasswordResetTokenInput) (*CreatePasswordResetTokenOutput, error) {
    // 1. バリデーション
    if err := uc.validator.Validate(ctx, validator.PasswordResetCreateValidatorInput{
        Email: input.Email,
    }); err != nil {
        return nil, err
    }

    // 2. ビジネスロジック + 永続化
    // ...
}
```

### Handler は UseCase を呼ぶだけ

```go
// internal/handler/sign_in_password/create.go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. UseCase 呼び出し（バリデーションはすべて UseCase 内で実行）
    output, err := h.authenticateByPasswordUC.Execute(ctx, usecase.AuthenticateByPasswordInput{
        Email:    email,
        Password: r.FormValue("password"),
    })
    if err != nil {
        // 2. エラー型に応じたレスポンス
        if formErrors := model.AsValidationError(err); formErrors != nil {
            // バリデーションエラー → フォームを再表示（セッションに保存してリダイレクト）
            if err := h.sessionMgr.SetValidationError(ctx, w, r, *formErrors); err != nil {
                slog.ErrorContext(ctx, "フォームエラーの設定に失敗", "error", err)
            }
            http.Redirect(w, r, "/sign_in/password", http.StatusSeeOther)
            return
        }
        // システムエラー → 500
        slog.ErrorContext(ctx, "パスワード認証に失敗しました", "error", err)
        h.sessionMgr.SetFlash(w, session.FlashError, i18n.T(ctx, "sign_in_error_server"))
        http.Redirect(w, r, "/sign_in/password", http.StatusSeeOther)
        return
    }

    // 3. 成功 → リダイレクト
    h.sessionMgr.SetSessionCookieByPublicID(w, r, output.PublicID)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

### DI の構築（main.go）

`main.go` で Validator → UseCase → Handler の順に構築します。

```go
// cmd/server/main.go での構築

// 1. Validator の構築（repository を注入）
signInPasswordValidator := validator.NewSignInPasswordCreateValidator(userRepo)

// 2. UseCase の構築（Validator を注入）
authenticateByPasswordUC := usecase.NewAuthenticateByPasswordUsecase(
    createSessionUC,
    signInPasswordValidator,
)

// 3. Handler の構築（UseCase を注入）
signInPasswordHandler := sign_in_password.NewHandler(cfg, sessionManager, authenticateByPasswordUC)
```

## テスト

### テスト方針

バリデーションのテストは `internal/validator/` パッケージ内の `*_test.go` ファイルに実装します。

| テスト対象                    | ファイル                       | 特徴                                 |
| ----------------------------- | ------------------------------ | ------------------------------------ |
| バリデーション                | `validator/*_test.go`          | 形式・状態バリデーションを統合テスト |
| UseCase（バリデーション含む） | `usecase/*_test.go`            | UseCase 経由のバリデーションテスト   |
| Handler の振る舞い            | `handler/{resource}/*_test.go` | E2E テスト、正常系・代表的な異常系   |

**理由**:

- **シンプルな構成**: バリデーターごとに 1 つのテストファイルにテストを集約
- **問題の特定**: テスト失敗時にどの検証の問題か即座に分かる
- **保守性向上**: テストファイルの管理が容易

### バリデーションのテスト

```go
// internal/validator/password_reset_test.go
func TestPasswordResetCreateValidator_Validate(t *testing.T) {
    t.Parallel()

    v := NewPasswordResetCreateValidator()

    tests := []struct {
        name       string
        input      PasswordResetCreateValidatorInput
        wantErrors bool
        wantField  string
    }{
        {
            name:       "有効な入力",
            input:      PasswordResetCreateValidatorInput{Email: "user@example.com"},
            wantErrors: false,
        },
        {
            name:       "メールアドレスが空",
            input:      PasswordResetCreateValidatorInput{Email: ""},
            wantErrors: true,
            wantField:  "email",
        },
        {
            name:       "スペースのみ",
            input:      PasswordResetCreateValidatorInput{Email: "   "},
            wantErrors: true,
            wantField:  "email",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            err := v.Validate(ctx, tt.input)
            formErrors := model.AsValidationError(err)

            if tt.wantErrors {
                if formErrors == nil {
                    t.Error("expected errors, got none")
                    return
                }
                if tt.wantField != "" && !formErrors.HasFieldError(tt.wantField) {
                    t.Errorf("expected field error for %q", tt.wantField)
                }
            } else {
                if formErrors != nil {
                    t.Errorf("unexpected errors: %v", formErrors)
                }
            }
        })
    }
}
```

### 状態バリデーションのテスト（DB 必要）

```go
// internal/validator/sign_in_password_test.go
func TestSignInPasswordCreateValidator_Validate(t *testing.T) {
    t.Parallel()

    db, tx := testutil.SetupTestDB(t)

    // テストユーザーを作成（パスワードは事前に bcrypt でハッシュ化）
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
    if err != nil {
        t.Fatalf("パスワードハッシュ生成エラー: %v", err)
    }
    testutil.NewUserBuilder(t, tx).
        WithEmail("test@example.com").
        WithEncryptedPassword(string(hashedPassword)).
        Build()

    userRepo := repository.NewUserRepository(query.New(db)).WithTx(tx)
    v := NewSignInPasswordCreateValidator(userRepo)

    t.Run("有効な認証情報", func(t *testing.T) {
        ctx := context.Background()
        output, err := v.Validate(ctx, SignInPasswordCreateValidatorInput{
            EmailOrUsername: "test@example.com",
            Password:        "password123",
        })
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if output == nil {
            t.Fatal("expected output on success")
        }
    })

    t.Run("無効なパスワード", func(t *testing.T) {
        ctx := context.Background()
        output, err := v.Validate(ctx, SignInPasswordCreateValidatorInput{
            EmailOrUsername: "test@example.com",
            Password:        "wrongpassword",
        })
        formErrors := model.AsValidationError(err)
        if formErrors == nil {
            t.Fatal("expected validation error")
        }
        if output != nil {
            t.Error("output should be nil on failure")
        }
        if len(formErrors.Global) == 0 {
            t.Error("expected global error for invalid credentials")
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

### 1. バリデーションは 1 ファイルに統合

```go
// ✅ Good: 状態バリデーションも internal/validator/ パッケージに配置
// internal/validator/sign_in_password.go
type SignInPasswordCreateValidator struct {
    userRepo *repository.UserRepository
}

func (v *SignInPasswordCreateValidator) Validate(ctx context.Context, input SignInPasswordCreateValidatorInput) (*SignInPasswordCreateValidateOutput, error) {
    // 1. 形式バリデーション（DB不要）
    ve := model.NewValidationError()
    // ...
    if ve.HasErrors() {
        return nil, ve
    }

    // 2. 状態バリデーション（DB必要）
    user, err := v.userRepo.GetByEmailOrUsername(ctx, input.EmailOrUsername)
    // ...
    return &SignInPasswordCreateValidateOutput{User: user}, nil
}
```

### 2. 国際化を徹底

```go
// ❌ Bad: ハードコードされたメッセージ
ve.AddField("email", "メールアドレスを入力してください")

// ✅ Good: 国際化された翻訳
ve.AddField("email", i18n.T(ctx, "password_reset_email_required"))
```

### 3. 早期リターンでネストを減らす

```go
// ❌ Bad: ネストが深い
func (v *PasswordResetCreateValidator) Validate(ctx context.Context, input PasswordResetCreateValidatorInput) error {
    ve := model.NewValidationError()
    if input.Email != "" {
        if emailRegex.MatchString(input.Email) {
            // OK
        } else {
            ve.AddField("email", "...")
        }
    } else {
        ve.AddField("email", "...")
    }
    if ve.HasErrors() {
        return ve
    }
    return nil
}

// ✅ Good: 早期リターンでシンプル
func (v *PasswordResetCreateValidator) Validate(ctx context.Context, input PasswordResetCreateValidatorInput) error {
    ve := model.NewValidationError()

    if strings.TrimSpace(input.Email) == "" {
        ve.AddField("email", i18n.T(ctx, "error_required"))
        return ve
    }

    if !emailRegex.MatchString(input.Email) {
        ve.AddField("email", i18n.T(ctx, "error_invalid_email_format"))
    }

    if ve.HasErrors() {
        return ve
    }
    return nil
}
```

### 4. 正規表現はパッケージレベルで定義

```go
// ✅ Good: 正規表現のコンパイルを1回だけ実行
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (v *PasswordResetCreateValidator) Validate(ctx context.Context, input PasswordResetCreateValidatorInput) error {
    // emailRegexを使用
}
```

### 5. Go の慣習に従った `(data, error)` の 2 値返し

```go
// ✅ Good: データを返す場合は (data, error)
// 成功時のデータは Output 構造体で返す
type SignInPasswordCreateValidateOutput struct {
    User repository.GetUserByEmailOrUsernameRow
}

// バリデーション失敗時は *model.ValidationError を error として返す
func (v *SignInPasswordCreateValidator) Validate(ctx context.Context, input SignInPasswordCreateValidatorInput) (*SignInPasswordCreateValidateOutput, error) {
    ve := model.NewValidationError()
    // ...
    if ve.HasErrors() {
        return nil, ve // *model.ValidationError は error を満たす
    }

    user, err := v.userRepo.GetByEmailOrUsername(ctx, input.EmailOrUsername)
    if err != nil {
        return nil, err // システムエラー
    }

    return &SignInPasswordCreateValidateOutput{User: user}, nil
}

// ✅ Good: データを返さない場合は error のみ
func (v *PasswordResetCreateValidator) Validate(ctx context.Context, input PasswordResetCreateValidatorInput) error {
    ve := model.NewValidationError()
    // ...
    if ve.HasErrors() {
        return ve
    }
    return nil
}
```

**判断基準**: Validator が DB から取得したデータを UseCase でも使う場合は Output 構造体を返す（二重クエリの回避）。使わない場合（失敗チェックだけ行う場合）は `error` のみ返す。

## 利点

1. **判断コストがゼロ**: バリデーションは常に `internal/validator/` に配置するため、「どこに書くべきか」を迷わない
2. **進化に強い**: 形式バリデーションに後から DB チェックが追加されても、ファイル移動が不要
3. **一箇所で把握できる**: バリデーションルールを確認したいとき `internal/validator/` だけ見ればよい
4. **アーキテクチャの強制**: Handler パッケージから `repository`・`validator` の import を完全に排除でき、depguard で強制可能
5. **依存が明確**: バリデーターの依存関係が一目でわかる
6. **テストしやすい**: Validator を独立してテストでき、UseCase テストではバリデーション込みの統合テストが可能
7. **再利用可能**: UseCase から呼び出されるため、エントリーポイント（Handler, Worker）が増えてもバリデーションが漏れない
