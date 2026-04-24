---
paths:
  - "go/**/*.{go,templ}"
---

# バリデーションガイド

このドキュメントは、Go 版 Wikino でのバリデーションのベストプラクティスを説明します。

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
├── sign_in.go            # バリデーション（形式チェック + DBを使った検証）
├── sign_in_test.go       # バリデーションのテスト
├── password_reset.go     # バリデーション（形式チェックのみ）
└── password_reset_test.go # バリデーションのテスト

internal/usecase/
├── create_sign_in.go     # UseCase（Validatorを外部から受け取り、呼び出す）
└── create_sign_in_test.go

internal/handler/sign_in/
├── handler.go            # Handler構造体と依存性（UseCaseを外部から受け取る）
├── new.go                # フォーム表示
└── create.go             # 作成処理
```

### バリデーションの分類

バリデーションは以下の 2 種類に分類されますが、同じファイル（`validator.go`）に実装します：

| 種類               | 責務                  | 特徴            |
| ------------------ | --------------------- | --------------- |
| 形式バリデーション | 入力値の形式チェック  | DB アクセス不要 |
| 状態バリデーション | DB の状態を使った検証 | DB アクセス必要 |

### 構造体の命名規則

- **命名規則**: `{Resource}{Action}Validator`（例: `SignInCreateValidator`, `PasswordResetCreateValidator`）
- **コンストラクタ**: `New{Resource}{Action}Validator`（例: `NewSignInCreateValidator`）
- **戻り値**: Go の慣習に従った `(data, error)` の 2 値返し。データを返す必要がない場合は `error` のみ
- **呼び出し元**: UseCase から呼び出される（Handler からの直接呼び出しは禁止）
- **1 つの構造体で両方のバリデーションを担当**: 形式バリデーションと状態バリデーションを `Validate` メソッド内で順次実行

### 状態バリデーションの配置場所

状態バリデーションは `internal/validator/` パッケージ内のバリデーターまたは UseCase のどちらかに配置します。

**判断基準**: **「検証失敗時に DB を更新する必要があるか？」**

| 検証失敗時の DB 更新 | 配置場所     | 理由                                               |
| -------------------- | ------------ | -------------------------------------------------- |
| 不要                 | validator.go | UseCase をシンプルに保つため                       |
| 必要                 | UseCase      | トランザクション内で検証と更新を行う必要があるため |

**validator.go で行うべき検証**:

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

※注: 「リカバリーコード消費」や「トークン使用済みマーク」は検証成功後の処理であり、バリデーションではなく UseCase の永続化処理として扱う。検証自体は validator.go で行い、成功後に UseCase を呼び出す。

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
// internal/validator/sign_in.go
package validator

import (
    "context"
    "net/mail"

    "github.com/wikinoapp/wikino/go/internal/auth"
    "github.com/wikinoapp/wikino/go/internal/i18n"
    "github.com/wikinoapp/wikino/go/internal/model"
    "github.com/wikinoapp/wikino/go/internal/repository"
)

// SignInCreateValidator はサインインのバリデーションを行う
type SignInCreateValidator struct {
    userRepo         *repository.UserRepository
    userPasswordRepo *repository.UserPasswordRepository
}

// NewSignInCreateValidator は SignInCreateValidator を生成する
func NewSignInCreateValidator(
    userRepo *repository.UserRepository,
    userPasswordRepo *repository.UserPasswordRepository,
) *SignInCreateValidator {
    return &SignInCreateValidator{
        userRepo:         userRepo,
        userPasswordRepo: userPasswordRepo,
    }
}

// SignInCreateValidatorInput はバリデーションの入力パラメータ
type SignInCreateValidatorInput struct {
    Email    string
    Password string
}

// SignInCreateValidateOutput はバリデーション成功時の出力
type SignInCreateValidateOutput struct {
    User *model.User
}

// Validate はバリデーションを行い、成功時はデータを返す
func (v *SignInCreateValidator) Validate(ctx context.Context, input SignInCreateValidatorInput) (*SignInCreateValidateOutput, error) {
    // 1. 形式バリデーション
    ve := model.NewValidationError()

    if input.Email == "" {
        ve.AddField("email", i18n.T(ctx, "validation_required"))
    } else if !isValidEmail(input.Email) {
        ve.AddField("email", i18n.T(ctx, "validation_email_invalid"))
    }

    if input.Password == "" {
        ve.AddField("password", i18n.T(ctx, "validation_required"))
    }

    if ve.HasErrors() {
        return nil, ve
    }

    // 2. 状態バリデーション（DB検証）
    user, err := v.userRepo.FindByEmail(ctx, input.Email)
    if err != nil {
        return nil, err
    }
    if user == nil {
        ve.AddGlobal(i18n.T(ctx, "validation_email_or_password_invalid"))
        return nil, ve
    }

    // パスワード検証
    userPassword, err := v.userPasswordRepo.FindByUserID(ctx, user.ID)
    if err != nil {
        return nil, err
    }
    if userPassword == nil || !auth.VerifyPassword(userPassword.PasswordDigest, input.Password) {
        ve.AddGlobal(i18n.T(ctx, "validation_email_or_password_invalid"))
        return nil, ve
    }

    return &SignInCreateValidateOutput{User: user}, nil
}

func isValidEmail(email string) bool {
    _, err := mail.ParseAddress(email)
    return err == nil
}
```

### データを返さないバリデーター（形式バリデーションのみ）

DB を使った検証が不要な場合でも、`internal/validator/` パッケージに配置します。データを返す必要がない場合は `error` のみを返します。

```go
// internal/validator/suggestion_comment.go
package validator

import (
    "context"
    "unicode/utf8"

    "github.com/wikinoapp/wikino/go/internal/i18n"
    "github.com/wikinoapp/wikino/go/internal/model"
)

const suggestionCommentBodyMaxLength = 10000

// SuggestionCommentCreateValidator は編集提案コメントのバリデーションを行う
type SuggestionCommentCreateValidator struct{}

// NewSuggestionCommentCreateValidator は SuggestionCommentCreateValidator を生成する
func NewSuggestionCommentCreateValidator() *SuggestionCommentCreateValidator {
    return &SuggestionCommentCreateValidator{}
}

// SuggestionCommentCreateValidatorInput はバリデーションの入力パラメータ
type SuggestionCommentCreateValidatorInput struct {
    Body string
}

// Validate はバリデーションを行う
func (v *SuggestionCommentCreateValidator) Validate(ctx context.Context, input SuggestionCommentCreateValidatorInput) error {
    ve := model.NewValidationError()

    if input.Body == "" {
        ve.AddField("body", i18n.T(ctx, "validation_suggestion_comment_body_required"))
    }

    if input.Body != "" && utf8.RuneCountInString(input.Body) > suggestionCommentBodyMaxLength {
        ve.AddField("body", i18n.T(ctx, "validation_suggestion_comment_body_too_long"))
    }

    if ve.HasErrors() {
        return ve
    }

    return nil
}
```

### 状態バリデーションで取得したデータを返すバリデーター

Validator は状態バリデーションの過程でデータを取得し、検証後にそのデータを戻り値として返します。これにより UseCase 内でデータを二重に取得する必要がなくなります。

```go
// internal/validator/suggestion.go
package validator

import (
    "context"
    "unicode/utf8"

    "github.com/wikinoapp/wikino/go/internal/i18n"
    "github.com/wikinoapp/wikino/go/internal/model"
    "github.com/wikinoapp/wikino/go/internal/repository"
)

// SuggestionCreateValidator は編集提案作成のバリデーションを行う
type SuggestionCreateValidator struct {
    draftPageRepo *repository.DraftPageRepository
}

// NewSuggestionCreateValidator は SuggestionCreateValidator を生成する
func NewSuggestionCreateValidator(draftPageRepo *repository.DraftPageRepository) *SuggestionCreateValidator {
    return &SuggestionCreateValidator{draftPageRepo: draftPageRepo}
}

// SuggestionCreateValidatorInput はバリデーションの入力パラメータ
type SuggestionCreateValidatorInput struct {
    Title         string
    Body          string
    DraftPageIDs  []model.DraftPageID
    SpaceMemberID model.SpaceMemberID
    TopicID       model.TopicID
    SpaceID       model.SpaceID
}

// Validate はバリデーションを行い、成功時はデータを返す
func (v *SuggestionCreateValidator) Validate(ctx context.Context, input SuggestionCreateValidatorInput) ([]*model.DraftPage, error) {
    ve := model.NewValidationError()

    if input.Title == "" {
        ve.AddField("title", i18n.T(ctx, "validation_suggestion_title_required"))
    }
    if input.Title != "" && utf8.RuneCountInString(input.Title) > suggestionTitleMaxLength {
        ve.AddField("title", i18n.T(ctx, "validation_suggestion_title_too_long"))
    }
    if len(input.DraftPageIDs) == 0 {
        ve.AddField("draft_page_ids", i18n.T(ctx, "validation_suggestion_draft_pages_required"))
    }

    if ve.HasErrors() {
        return nil, ve
    }

    // 状態バリデーションで取得したデータを返す
    draftPages, err := v.draftPageRepo.FindByIDs(ctx, input.DraftPageIDs, input.SpaceID)
    if err != nil {
        return nil, err  // システムエラー
    }

    return draftPages, nil
}
```

## UseCase での使用

Validator は UseCase から呼び出されます。Handler は UseCase を呼び出すだけで、Validator に直接依存しません。

### UseCase でのバリデーション呼び出し

```go
// internal/usecase/create_suggestion_comment.go
func (uc *CreateSuggestionCommentUsecase) Execute(ctx context.Context, input CreateSuggestionCommentInput) (*CreateSuggestionCommentOutput, error) {
    // 1. データ取得
    space, spaceMember, suggestion, err := uc.fetchData(ctx, input)
    if err != nil {
        return nil, err
    }

    // 2. 認可チェック（Policy）
    if err := uc.authorize(ctx, space, spaceMember, suggestion); err != nil {
        return nil, err
    }

    // 3. バリデーション（Validator を呼び出す）
    if err := uc.createValidator.Validate(ctx, validator.SuggestionCommentCreateValidatorInput{
        Body: input.Body,
    }); err != nil {
        return nil, err  // *model.ValidationError か素の error がそのまま上がる
    }

    // 4. ビジネスロジック + 永続化
    return uc.createComment(ctx, space.ID, suggestion.ID, spaceMember.ID, input.Body)
}
```

### Handler は UseCase を呼ぶだけ

```go
// internal/handler/suggestion_comment/create.go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. リクエストのパース
    body := r.FormValue("body")

    // 2. UseCase 呼び出し（認可・バリデーションはすべて UseCase 内で実行）
    _, err := h.createSuggestionCommentUsecase.Execute(ctx, usecase.CreateSuggestionCommentInput{
        Body: body,
        // ...
    })
    if err != nil {
        // エラー型に応じたレスポンス
        if ve := model.AsValidationError(err); ve != nil {
            // バリデーションエラー → フラッシュメッセージでリダイレクト
            errs := ve.GetFieldErrors("body")
            if len(errs) > 0 {
                h.flashMgr.SetError(w, errs[0])
            }
            http.Redirect(w, r, suggestionPath, http.StatusSeeOther)
            return
        }
        if ae := model.AsAppError(err); ae != nil {
            handler.NotFound(w, r)
            return
        }
        slog.ErrorContext(ctx, "編集提案コメントの作成に失敗", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // 3. レスポンス
    http.Redirect(w, r, suggestionPath, http.StatusSeeOther)
}
```

### DI の構築（main.go）

`main.go` で Validator → UseCase → Handler の順に構築します。

```go
// cmd/server/main.go での構築

// 1. Validator の構築
suggestionCommentCreateValidator := validator.NewSuggestionCommentCreateValidator()

// 2. UseCase の構築（Validator を注入）
createSuggestionCommentUC := usecase.NewCreateSuggestionCommentUsecase(
    db, spaceRepo, spaceMemberRepo, topicMemberRepo,
    suggestionRepo, suggestionCommentRepo,
    suggestionCommentCreateValidator,
)

// 3. Handler の構築（UseCase を注入）
suggestionCommentHandler := suggestion_comment.NewHandler(flashMgr, createSuggestionCommentUC)
```

## テスト

### テスト方針

バリデーションのテストは `validator_test.go` に統合して実装します。

| テスト対象           | ファイル            | 特徴                                   |
| -------------------- | ------------------- | -------------------------------------- |
| バリデーション全体   | `validator_test.go` | 形式・状態バリデーションを統合テスト   |
| ハンドラーの振る舞い | `handler_test.go`   | E2E テスト、正常系・代表的な異常系のみ |

**理由**:

- **シンプルな構成**: 1 つのファイルにすべてのバリデーションテストを集約
- **問題の特定**: テスト失敗時にどの検証の問題か即座に分かる
- **保守性向上**: テストファイルの管理が容易

### バリデーションのテスト

```go
// internal/validator/suggestion_comment_test.go
func TestSuggestionCommentCreateValidator_Validate(t *testing.T) {
    t.Parallel()

    v := NewSuggestionCommentCreateValidator()

    tests := []struct {
        name          string
        input         SuggestionCommentCreateValidatorInput
        wantErr       bool
        expectedField string
    }{
        {
            name:    "正常系: 有効な入力",
            input:   SuggestionCommentCreateValidatorInput{Body: "コメント本文"},
            wantErr: false,
        },
        {
            name:          "異常系: 本文が空",
            input:         SuggestionCommentCreateValidatorInput{Body: ""},
            wantErr:       true,
            expectedField: "body",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := testutil.ContextWithLocale(t, "ja")

            err := v.Validate(ctx, tt.input)

            if tt.wantErr {
                ve := model.AsValidationError(err)
                if ve == nil {
                    t.Fatal("expected ValidationError, got nil")
                }
                if tt.expectedField != "" {
                    errs := ve.GetFieldErrors(tt.expectedField)
                    if len(errs) == 0 {
                        t.Errorf("expected field error for %q", tt.expectedField)
                    }
                }
            } else {
                if err != nil {
                    t.Fatalf("unexpected error: %v", err)
                }
            }
        })
    }
}
```

### 状態バリデーションのテスト（DB 必要）

```go
// internal/validator/sign_in_test.go
func TestSignInCreateValidator_Validate(t *testing.T) {
    t.Parallel()

    t.Run("正常系: 有効な認証情報", func(t *testing.T) {
        t.Parallel()
        db, tx := testutil.SetupTx(t)

        // テストデータの作成
        userID := testutil.NewUserBuilder(t, tx).
            WithEmail("test@example.com").
            WithPassword("password123").
            Build()

        userRepo := repository.NewUserRepository(db).WithTx(tx)
        userPasswordRepo := repository.NewUserPasswordRepository(db).WithTx(tx)
        v := NewSignInCreateValidator(userRepo, userPasswordRepo)

        ctx := testutil.ContextWithLocale(t, "ja")
        output, err := v.Validate(ctx, SignInCreateValidatorInput{
            Email:    "test@example.com",
            Password: "password123",
        })

        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if output.User == nil {
            t.Fatal("expected user, got nil")
        }
    })

    t.Run("異常系: 無効なパスワード", func(t *testing.T) {
        t.Parallel()
        db, tx := testutil.SetupTx(t)

        testutil.NewUserBuilder(t, tx).
            WithEmail("test@example.com").
            WithPassword("password123").
            Build()

        userRepo := repository.NewUserRepository(db).WithTx(tx)
        userPasswordRepo := repository.NewUserPasswordRepository(db).WithTx(tx)
        v := NewSignInCreateValidator(userRepo, userPasswordRepo)

        ctx := testutil.ContextWithLocale(t, "ja")
        output, err := v.Validate(ctx, SignInCreateValidatorInput{
            Email:    "test@example.com",
            Password: "wrongpassword",
        })

        if output != nil {
            t.Error("expected nil output")
        }
        ve := model.AsValidationError(err)
        if ve == nil {
            t.Fatal("expected ValidationError")
        }
        if !ve.HasErrors() {
            t.Error("expected validation errors")
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
// ✅ Good: 状態バリデーションは internal/validator/ パッケージに配置
// internal/validator/sign_in.go
type SignInCreateValidator struct {
    userRepo         *repository.UserRepository
    userPasswordRepo *repository.UserPasswordRepository
}

func (v *SignInCreateValidator) Validate(ctx context.Context, input SignInCreateValidatorInput) (*SignInCreateValidateOutput, error) {
    // 1. 形式バリデーション（DB不要）
    ve := model.NewValidationError()
    // ...
    if ve.HasErrors() {
        return nil, ve
    }

    // 2. 状態バリデーション（DB必要）
    user, err := v.userRepo.FindByEmail(ctx, input.Email)
    // ...
    return &SignInCreateValidateOutput{User: user}, nil
}
```

### 2. 国際化を徹底

```go
// ❌ Bad: ハードコードされたメッセージ
ve.AddField("email", "メールアドレスを入力してください")

// ✅ Good: 国際化された翻訳
ve.AddField("email", i18n.T(ctx, "error_required"))
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

    if input.Email == "" {
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
func (v *SuggestionCreateValidator) Validate(ctx context.Context, input SuggestionCreateValidatorInput) ([]*model.DraftPage, error) {
    ve := model.NewValidationError()
    // ...
    if ve.HasErrors() {
        return nil, ve  // *model.ValidationError は error を満たす
    }

    draftPages, err := v.draftPageRepo.FindByIDs(ctx, input.DraftPageIDs, input.SpaceID)
    if err != nil {
        return nil, err  // システムエラー
    }

    return draftPages, nil
}

// ✅ Good: データを返さない場合は error のみ
func (v *SuggestionCommentCreateValidator) Validate(ctx context.Context, input SuggestionCommentCreateValidatorInput) error {
    ve := model.NewValidationError()
    // ...
    if ve.HasErrors() {
        return ve
    }
    return nil
}
```

## 利点

1. **判断コストがゼロ**: バリデーションは常に `internal/validator/` に配置するため、「どこに書くべきか」を迷わない
2. **進化に強い**: 形式バリデーションに後から DB チェックが追加されても、ファイル移動が不要
3. **一箇所で把握できる**: バリデーションルールを確認したいとき `internal/validator/` だけ見ればよい
4. **アーキテクチャの強制**: Handler パッケージから repository・validator の import を完全に排除でき、depguard で強制可能
5. **依存が明確**: バリデーターの依存関係が一目でわかる
6. **テストしやすい**: バリデーション全体を独立してテストできる
7. **再利用可能**: UseCase から呼び出されるため、エントリーポイント（Handler, Worker）が増えても認可・バリデーションが漏れない
