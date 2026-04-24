---
paths:
  - "go/**/*.{go,templ}"
---

# ユースケースガイド

このドキュメントは、Go 版 Wikino の UseCase 層の設計と実装パターンを説明します。

## 概要

UseCase はアプリケーションのオーケストレーターです。Handler/Worker からのすべてのデータアクセス・認可・バリデーション・ビジネスロジック・永続化は UseCase を経由します。

## 責務

- **オーケストレーション**: 認可チェック・バリデーション・ビジネスロジック・永続化を統括する
- データ取得ロジックの集約（読み取り UseCase）
- トランザクション管理（書き込み UseCase: `db.BeginTx` から `tx.Commit` まで）
- 複数の repository を跨ぐ処理

## UseCase の役割

Handler/Worker からのすべてのデータアクセスは UseCase を経由する。UseCase は読み取りと書き込みの両方を担当する。

| 種類             | 責務                                                 | トランザクション        |
| ---------------- | ---------------------------------------------------- | ----------------------- |
| 書き込み UseCase | 認可・バリデーション・ビジネスロジック・永続化を統括 | あり（WithTx パターン） |
| 読み取り UseCase | データ取得、複数 Repository の集約                   | なし                    |

**書き込み UseCase**:

- 認可チェック・バリデーション・ビジネスロジック・永続化を統括するオーケストレーター
- トランザクションを伴う永続化処理（作成・更新・削除）
- 複数の Repository を跨ぐビジネスロジック
- ロールバックが必要な複合操作
- **書き込み UseCase のルール**（詳細は「UseCase 内の処理順序」を参照）:
  1. トランザクション開始後はデータの取得や計算処理を行わない。永続化処理のみ行う（トランザクション前のデータ取得は許可）
  2. Execute 内にロジックを直接書かない。ロジックは関数やメソッドとして定義し、Execute 内ではそれを呼び出すだけにする

```go
// 例: ページとスペースメンバーを同時に更新する場合
type CreatePageUsecase struct {
    db              *sql.DB
    pageRepo        *repository.PageRepository
    spaceMemberRepo *repository.SpaceMemberRepository
}

func (uc *CreatePageUsecase) Execute(ctx context.Context, input Input) (*Result, error) {
    tx, err := uc.db.BeginTx(ctx, nil)
    // トランザクション内で複数のRepositoryを操作
}
```

**読み取り UseCase**:

- Handler が必要とするデータ取得ロジックを集約する
- 複数の Repository を組み合わせてデータを取得する
- トランザクションは不要

```go
// 例: トピック詳細ページのデータ取得
type GetTopicDetailUsecase struct {
    spaceRepo       *repository.SpaceRepository
    spaceMemberRepo *repository.SpaceMemberRepository
    topicRepo       *repository.TopicRepository
    topicMemberRepo *repository.TopicMemberRepository
    pageRepo        *repository.PageRepository
}

type GetTopicDetailInput struct {
    SpaceIdentifier model.SpaceIdentifier
    TopicNumber     int32
    UserID          *model.UserID
    Page            int32
}

type GetTopicDetailOutput struct {
    Space       *model.Space
    SpaceMember *model.SpaceMember
    Topic       *model.Topic
    TopicMember *model.TopicMember
    PinnedPages []*model.Page
    Pages       []*model.Page
    Pagination  *repository.PaginationResult
}

func (uc *GetTopicDetailUsecase) Execute(ctx context.Context, input GetTopicDetailInput) (*GetTopicDetailOutput, error) {
    space, err := uc.spaceRepo.FindByIdentifier(ctx, input.SpaceIdentifier)
    if err != nil {
        return nil, fmt.Errorf("スペースの取得に失敗: %w", err)
    }
    // 複数のRepositoryからデータを取得して集約
    // ...
}
```

## ファイル配置

`internal/usecase/` 直下にフラットに配置（サブディレクトリは作成しない）

**プライベート関数の配置ルール**: あるUseCaseファイルに定義されたプライベート関数を別のUseCaseファイルから呼び出す必要が生じた場合は、その関数を専用のファイルに切り出す。ファイル名は関数の責務を表す名詞にする（例: Wikiリンク関連の共通関数を `linked_page.go` に配置）。

## 命名規則

- **ファイル名**: `{action}_{entity}.go`
  - 例: `create_session.go`, `create_password_reset_token.go`, `update_password_reset.go`
  - **重要**: 動詞（アクション）を必ず先頭に配置する
- **構造体名**: `{Action}{Entity}Usecase`
  - 例: `CreateSessionUsecase`, `CreatePasswordResetTokenUsecase`
  - 注: `Usecase` の `c` は小文字（既存コードとの統一のため）
- **読み取り UseCase のプレフィックスは `Get` に統一**: 読み取り UseCase のアクションには `Get` を使用する。`List` や `Fetch` など他の動詞は使用しない。これにより `Get` = 読み取り、それ以外 = 書き込みという判別が即座にできる
  - 例: `GetTopicDetailUsecase`, `GetPageDetailUsecase`, `GetDraftPagesUsecase`
- **コンストラクタ**: `New{Action}{Entity}Usecase`
- **Execute メソッド**: `Execute(ctx context.Context, ...) (*Result, error)`

## 結果型

各 UseCase は専用の Result 構造体を返します。

例: `SessionResult`, `CreatePasswordResetTokenResult`

## 利点

- ハンドラーがシンプルになる（HTTP 処理に専念できる）
- トランザクション境界が明確
- テストしやすい構造
- ビジネスロジックの再利用が可能

## UseCase 内の処理順序

書き込み UseCase は以下の順序で処理を実行する。

```
書き込み UseCase (オーケストレーター)
  1. データ取得（トランザクション外）
  2. 認可チェック（Policy）
  3. バリデーション（Validator）
  4. ビジネスロジック（計算、変換等）
  5. トランザクション（永続化のみ）
```

### 書き込み UseCase のルール

書き込み UseCase は以下の 2 つのルールを守る:

1. **トランザクション開始後はデータの取得や計算処理を行わない**: トランザクション内は永続化処理のみ行う。ただし、トランザクション開始前であればデータの取得や計算処理を行ってよい
2. **Execute 内にロジックを直接書かない**: ロジックは関数やメソッドとして定義し、Execute 内ではそれを呼び出すだけにする

```go
// ✅ 良い例: 書き込み UseCase がデータ取得・認可・バリデーション・永続化を統括する
func (uc *CreateSuggestionUsecase) Execute(ctx context.Context, input CreateSuggestionInput) (*CreateSuggestionOutput, error) {
    // 1. データ取得（トランザクション外）
    space, err := uc.spaceRepo.FindByIdentifier(ctx, input.SpaceIdentifier)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            return nil, &model.AppError{
                Code:     model.AppErrCodeResourceNotFound,
                UserMsg:  i18n.T(ctx, "error_space_not_found"),
                Internal: err,
            }
        }
        return nil, err
    }

    // 2. 認可チェック
    if !policy.NewTopicPolicy(spaceMember, topicMember).CanCreateSuggestion() {
        return nil, &model.AppError{
            Code:    model.AppErrCodeForbidden,
            UserMsg: i18n.T(ctx, "error_forbidden"),
        }
    }

    // 3. バリデーション
    draftPages, err := uc.validator.Validate(ctx, validatorInput)
    if err != nil {
        return nil, err  // *model.ValidationError か素の error がそのまま上がる
    }

    // 4. ビジネスロジック + 5. トランザクション（永続化のみ）
    return uc.createSuggestion(ctx, input, draftPages)
}

// ❌ 悪い例: トランザクション内でデータ取得を行っている
func (uc *WriteUsecase) Execute(ctx context.Context, input Input) error {
    tx, err := uc.db.BeginTx(ctx, nil)
    // トランザクション内でデータ取得 → トランザクション前に行うべき
    page, err := pageRepo.FindByID(ctx, input.PageID, input.SpaceID)
    // ...
}
```

### エラー型の使い分け

UseCase は以下の 3 種類のエラーを返す。Handler は `errors.As` でエラーの型を判別してレスポンスを決定する。

| エラー型                 | 生成元    | 意味                             | Handler の対応                          |
| ------------------------ | --------- | -------------------------------- | --------------------------------------- |
| `*model.ValidationError` | Validator | 入力が不正（ユーザーが修正可能） | フォーム再描画（422）                   |
| `*model.AppError`        | UseCase   | 業務レベルの既知の失敗           | エラーコードに応じた処理（403, 404 等） |
| 素の `error`             | どこでも  | 予期しないシステムエラー         | 500                                     |

### Handler の実装パターン

Handler は薄い Adapter として、リクエストのパース → UseCase 呼び出し → レスポンス生成のみを行う。

```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. リクエストのパース
    title := r.FormValue("title")
    // ...

    // 2. UseCase 呼び出し（認可・バリデーション・永続化はすべて UseCase 内で実行）
    output, err := h.createSuggestionUC.Execute(ctx, usecase.CreateSuggestionInput{
        Title:           title,
        SpaceIdentifier: spaceIdentifier,
        UserID:          userID,
    })
    if err != nil {
        var ve *model.ValidationError
        if errors.As(err, &ve) {
            // バリデーションエラー → フォーム再描画（422）
            w.WriteHeader(http.StatusUnprocessableEntity)
            h.renderNewForm(w, r, ve)
            return
        }
        var ae *model.AppError
        if errors.As(err, &ae) {
            // アプリケーションエラー → ログ + エラーコードに応じた処理
            slog.ErrorContext(ctx, ae.LogString())
            h.renderError(w, r, ae)
            return
        }
        // 予期しないエラー → 500
        slog.ErrorContext(ctx, "予期しないエラー", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // 3. レスポンス
    http.Redirect(w, r, redirectPath, http.StatusSeeOther)
}
```

**Validator でのデータ取得パターン**: Validator は状態バリデーションの過程でデータを取得し、検証後にそのデータを戻り値として返す。これにより UseCase 内でデータを二重に取得する必要がなくなる。Validator は Go の慣習に従った `(data, error)` の 2 値返しを使用する。

```go
// Validator が検証の過程で取得したデータを戻り値として返す
func (v *SuggestionCreateValidator) Validate(ctx context.Context, input Input) ([]*model.DraftPage, error) {
    ve := model.NewValidationError()

    if input.Title == "" {
        ve.AddField("title", templates.T(ctx, "error_required"))
    }
    if ve.HasErrors() {
        return nil, ve  // *model.ValidationError は error を満たす
    }

    // 状態バリデーションで取得したデータを返す
    draftPages, err := v.draftPageRepo.FindByIDs(ctx, input.DraftPageIDs)
    if err != nil {
        return nil, err  // システムエラー
    }

    return draftPages, nil
}
```

## 実装例

### シンプルなユースケース（トランザクションなし）

```go
// internal/usecase/create_session.go
package usecase

import (
    "context"
    "github.com/wikinoapp/wikino/internal/repository"
)

type CreateSessionUsecase struct {
    queries *repository.Queries
}

func NewCreateSessionUsecase(queries *repository.Queries) *CreateSessionUsecase {
    return &CreateSessionUsecase{queries: queries}
}

type SessionResult struct {
    PublicID string
    UserID   int64
}

func (uc *CreateSessionUsecase) Execute(ctx context.Context, userID int64) (*SessionResult, error) {
    // セッションIDを生成
    publicID := generateSecureRandomString(32)

    // セッションをDBに保存
    session, err := uc.queries.CreateSession(ctx, repository.CreateSessionParams{
        PublicID: publicID,
        UserID:   userID,
    })
    if err != nil {
        return nil, fmt.Errorf("セッションの作成に失敗: %w", err)
    }

    return &SessionResult{
        PublicID: session.PublicID,
        UserID:   session.UserID,
    }, nil
}
```

### 複雑なユースケース（トランザクションあり）

```go
// internal/usecase/create_password_reset_token.go
package usecase

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    "github.com/wikinoapp/wikino/internal/repository"
)

type CreatePasswordResetTokenUsecase struct {
    db      *sql.DB
    queries *repository.Queries
}

func NewCreatePasswordResetTokenUsecase(db *sql.DB, queries *repository.Queries) *CreatePasswordResetTokenUsecase {
    return &CreatePasswordResetTokenUsecase{
        db:      db,
        queries: queries,
    }
}

type CreatePasswordResetTokenResult struct {
    Token  string
    UserID int64
}

func (uc *CreatePasswordResetTokenUsecase) Execute(ctx context.Context, userID int64) (*CreatePasswordResetTokenResult, error) {
    // トランザクション開始
    tx, err := uc.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("トランザクションの開始に失敗: %w", err)
    }
    defer tx.Rollback()

    // トランザクション対応のクエリを作成
    qtx := uc.queries.WithTx(tx)

    // 既存のトークンを削除
    err = qtx.DeletePasswordResetTokensByUserID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("既存トークンの削除に失敗: %w", err)
    }

    // 新しいトークンを生成
    token := generateSecureRandomString(32)
    hashedToken := hashToken(token)

    // トークンをDBに保存
    expiresAt := time.Now().Add(24 * time.Hour)
    _, err = qtx.CreatePasswordResetToken(ctx, repository.CreatePasswordResetTokenParams{
        UserID:      userID,
        Token:       hashedToken,
        ExpiresAt:   expiresAt,
    })
    if err != nil {
        return nil, fmt.Errorf("トークンの作成に失敗: %w", err)
    }

    // トランザクションをコミット
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
    }

    return &CreatePasswordResetTokenResult{
        Token:  token,
        UserID: userID,
    }, nil
}
```

## ハンドラーでの使用

Handler は薄い Adapter として UseCase を呼び出すだけです。具体的な実装パターンは「[UseCase 内の処理順序](#usecase-内の処理順序)」セクションの「Handler の実装パターン」を参照してください。

## Repository の WithTx パターン

Usecase でトランザクションを使用する場合、**Repository の `WithTx` メソッド**を使用してトランザクション内で操作するリポジトリを取得します。

### なぜ WithTx パターンを使うのか

**メリット**:

- **依存性注入**: Repository をコンストラクタで受け取るため、テストでモックに差し替えやすい
- **意図が明確**: `WithTx(tx)` の呼び出しで「このリポジトリはトランザクション内で操作する」という意図が明確
- **一貫性**: すべての Usecase で同じパターンを使用することで、コードの読みやすさが向上

### Repository に WithTx を実装する

各 Repository には `WithTx` メソッドを実装します：

```go
// internal/repository/user_repository.go

// WithTx はトランザクションを使用する新しいRepositoryを返す
func (r *UserRepository) WithTx(tx *sql.Tx) *UserRepository {
    return &UserRepository{q: r.q.WithTx(tx)}
}
```

### Usecase で WithTx を使用する

```go
// internal/usecase/create_account.go

type CreateAccountUsecase struct {
    db                    *sql.DB
    emailConfirmationRepo *repository.EmailConfirmationRepository
    userRepo              *repository.UserRepository
    userPasswordRepo      *repository.UserPasswordRepository
}

func NewCreateAccountUsecase(
    db *sql.DB,
    emailConfirmationRepo *repository.EmailConfirmationRepository,
    userRepo *repository.UserRepository,
    userPasswordRepo *repository.UserPasswordRepository,
) *CreateAccountUsecase {
    return &CreateAccountUsecase{
        db:                    db,
        emailConfirmationRepo: emailConfirmationRepo,
        userRepo:              userRepo,
        userPasswordRepo:      userPasswordRepo,
    }
}

func (uc *CreateAccountUsecase) Execute(ctx context.Context, input CreateAccountInput) (*CreateAccountOutput, error) {
    // トランザクションを開始
    tx, err := uc.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
    }
    defer func() {
        _ = tx.Rollback()
    }()

    // トランザクション内で操作するためのリポジトリを取得
    emailConfirmationRepo := uc.emailConfirmationRepo.WithTx(tx)
    userRepo := uc.userRepo.WithTx(tx)
    userPasswordRepo := uc.userPasswordRepo.WithTx(tx)

    // 以降の処理はトランザクション内のリポジトリを使用
    // ...

    // トランザクションをコミット
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
    }

    return &CreateAccountOutput{UserID: user.ID}, nil
}
```

### 重要なポイント

1. **Repository はコンストラクタで受け取る**: `NewXxxUsecase` で Repository を引数として受け取る
2. **Execute 内で WithTx を呼び出す**: トランザクションを開始した後、各 Repository の `WithTx(tx)` を呼び出す
3. **元の Repository は変更しない**: `WithTx` は新しい Repository インスタンスを返すため、元の Repository には影響しない
4. **すべての Repository で WithTx を使う**: トランザクション内で使用するすべての Repository に対して `WithTx` を呼び出す

## テスト

```go
func TestCreatePasswordResetTokenUsecase_Execute(t *testing.T) {
    // 共有DB接続プールからトランザクションをセットアップ
    db, tx := testutil.SetupTx(t)
    queries := repository.New(db).WithTx(tx)

    // テストユーザーを作成
    userID := testutil.NewUserBuilder(t, tx).
        WithEmail("test@example.com").
        Build()

    // ユースケースを実行
    uc := usecase.NewCreatePasswordResetTokenUsecase(db, queries)
    result, err := uc.Execute(context.Background(), userID)

    // アサーション
    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    if result.Token == "" {
        t.Error("Token should not be empty")
    }

    if result.UserID != userID {
        t.Errorf("UserID = %d, want %d", result.UserID, userID)
    }

    // トークンがDBに保存されているか確認
    tokens, err := queries.GetPasswordResetTokensByUserID(context.Background(), userID)
    if err != nil {
        t.Fatalf("GetPasswordResetTokensByUserID() error = %v", err)
    }

    if len(tokens) != 1 {
        t.Errorf("len(tokens) = %d, want 1", len(tokens))
    }
}
```

## 命名の注意点

### ファイル名の順序

`{action}_{entity}` の順（動詞を必ず先頭に）

- ✅ `create_session.go`
- ❌ `session_create.go`
- ✅ `create_password_reset_token.go`
- ❌ `password_reset_create_token.go`

### 複合エンティティ

エンティティが複数単語の場合はアンダースコアで連結

- ✅ `create_password_reset_token.go` （password_reset_token というエンティティ）

### 構造体名の大文字化

`Usecase` の `c` は小文字

- ✅ `CreateSessionUsecase`
- ❌ `CreateSessionUseCase`

## ファイル配置の理由

### フラット構造

エンティティごとにディレクトリを作らず、`internal/usecase/` 直下にすべてのファイルを配置

理由:

- **検索性**: ファイル名のプレフィックスでグルーピングされるため、エディタで検索しやすい
- **シンプルさ**: ディレクトリ階層が深くならず、import パスがシンプル
- **スケーラビリティ**: ファイル数が増えても管理しやすい

## 採用しなかった方針

### A. 書き込み UseCase のために読み取り UseCase を新設する

書き込み UseCase からすべてのデータ取得を外出しし、書き込み UseCase のためだけに読み取り UseCase を作成する方針。

**不採用の理由**:

- Handler が書き込み UseCase の内部実装を知る必要が生じる（どんなデータを事前に用意すべきか）
- 書き込み UseCase のために読み取り UseCase を作ると、両者が強く結合し、分離のメリットが薄い
- 命名が酷似し混同しやすくなる（例: `GetDraftPageSaveDataUsecase` と `GetSaveDraftPageDataUsecase`）

**代替として採用した方針**: 書き込み UseCase 内であっても、トランザクション開始前であればデータ取得を行ってよい。書き込み UseCase のルール（トランザクション内は永続化のみ、Execute 内にロジックを直接書かない）を守る限り、データ取得の配置場所は柔軟に判断する。

### B. Handler がオーケストレーターとして認可・バリデーションを制御する

Handler が読み取り UseCase → Policy → Validator → 書き込み UseCase の流れを制御する方針。

**不採用の理由**:

- エントリーポイントが増えた場合（Web API など）、認可・バリデーションの呼び出しを各エントリーポイントで再現する必要があり、漏れが発生しやすい
- 外部世界との接点である Handler にビジネスロジックの制御フローが書かれており、関心の分離が不十分
- Handler にドメイン固有の判断が集中し、テストが複雑になる

**代替として採用した方針**: UseCase をオーケストレーターにする。バリデーション・認可・ビジネスロジック・永続化を UseCase 内部で統括し、Handler は HTTP の入出力変換に徹する。

### C. Read UseCase を廃止し UseCase を1つに統合する

GET と POST で同じ UseCase を呼び、引数で動作を切り替える方針。

**不採用の理由**:

- GET（フォーム表示）と POST（作成処理）で責務が異なるため、1つの UseCase に統合すると不自然になる
- 読み取り UseCase はフォーム表示専用として残すほうが、責務が明確でシンプル
