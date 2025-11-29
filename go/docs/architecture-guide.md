# アーキテクチャガイド

このドキュメントは、Go版Annictのアーキテクチャパターンを説明します。

## 概要

Go版Annictは、関心の分離を意識した**3層アーキテクチャ**を採用しています。

### 3層アーキテクチャの構成

```
┌─────────────────────────────────────────────────────────┐
│ Presentation層（プレゼンテーション層）                    │
│ - Handler                                              │
│ - ViewModel                                            │
│ - Template                                             │
│ - Middleware                                           │
│ - Presentation層のヘルパー（image, i18n, session）       │
└─────────────────────────────────────────────────────────┘
         ↓ 依存（OK）
┌─────────────────────────────────────────────────────────┐
│ Application層（アプリケーション層）                        │
│ - UseCase（ビジネスフロー、トランザクション管理）           │
└─────────────────────────────────────────────────────────┘
         ↓ 依存（OK）
┌─────────────────────────────────────────────────────────┐
│ Domain/Infrastructure層（統合）                          │
│ - Query (sqlc)                                         │
│ - Repository                                           │
│ - Model                                                │
│ （同じ層なので相互に依存できる）                          │
└─────────────────────────────────────────────────────────┘
```

**重要**: Domain/Infrastructure層はPresentation層に**依存してはいけない**

### Domain/Infrastructure層を統合する理由

このプロジェクトでは、Domain層とInfrastructure層を分離せず、統合して扱います：

- **実用的**: データベース変更（PostgreSQL → MySQLなど）はほぼ起こらない
- **シンプル**: 層をまたぐ変換コストを削減し、シンプルさを保つ
- **Goらしい**: Goのプラグマティックな哲学に合致
- **YAGNI原則**: 必要になってから層を分ければ良い

RepositoryとModelを同じ層として扱うことで、依存関係がシンプルになり、相互に依存できます。

### ModelとRepositoryの1:1関係

各ドメインエンティティに対して対応するRepositoryを作成します：

- `model.Work` ↔ `repository.WorkRepository`
- `model.User` ↔ `repository.UserRepository`
- `model.Episode` ↔ `repository.EpisodeRepository`

この1:1関係により、以下のメリットがあります：

- **一貫性**: どのModelに対してどのRepositoryを使うかが明確
- **保守性**: Modelの変更がRepositoryに集約される
- **可読性**: コードの見通しが良くなる

### データの流れ

1. **Query** (Domain/Infrastructure層): SQLクエリを実行し、クエリ結果（`query.GetPopularWorksRow`など）を返す
2. **Repository** (Domain/Infrastructure層): Query結果をModelに変換し、複数のクエリを組み合わせる
3. **Model** (Domain/Infrastructure層): ページに依存しない汎用的なドメインエンティティ（`model.Work`など）
4. **Handler** (Presentation層): RepositoryからModelを取得し、ModelをViewModelに変換
5. **ViewModel** (Presentation層): 表示用のデータ構造（画像URL生成、言語切り替えなど）
6. **Template** (Presentation層): ViewModelを受け取ってHTMLを生成

### 主要なレイヤー

- **Handler**: HTTPリクエスト・レスポンスの処理
- **UseCase**: ビジネスロジックとトランザクション管理
- **ViewModel**: プレゼンテーション層のデータ変換
- **Repository**: Query結果をModelに変換
- **Model**: ドメインエンティティ
- **Query**: sqlc生成コード（データアクセス層）

## レイヤーごとのパッケージ分類

### Presentation層（プレゼンテーション層）

- **internal/handler**: HTTP リクエストハンドラー（リソースごとにディレクトリを切り、1 エンドポイント = 1 ファイルの原則）
- **internal/viewmodel**: プレゼンテーション層のデータ変換（View 用のデータ構造）
- **internal/templates**: templ テンプレートファイル（型安全な HTML テンプレート）
  - `layouts/`: レイアウトテンプレート（`default.templ`, `simple.templ`）
  - `components/`: 再利用可能なコンポーネント（`head.templ`, `flash.templ`, `form_errors.templ`）
  - `pages/`: ページテンプレート（機能別にディレクトリを分割）
  - `emails/`: メールテンプレート
  - `helper.go`: テンプレートヘルパー関数（`T()`, `Locale()`, `Icon()`など）
- **internal/middleware**: HTTP ミドルウェア
  - `reverse_proxy.go`: Rails 版へのリバースプロキシミドルウェア（Go 版で未実装の機能を Rails 版にプロキシ）
  - `auth.go`: 認証ミドルウェア
  - `csrf.go`: CSRF 保護ミドルウェア
  - `method_override.go`: HTTP メソッドオーバーライドミドルウェア

**Presentation層のヘルパー**（Presentation層内で使用可能）:

- **internal/image**: 画像URL生成（imgproxy署名付きURL）
- **internal/i18n**: 国際化（翻訳取得、言語切り替え）
- **internal/session**: セッション管理（フラッシュメッセージ、ユーザー情報）

### Application層（アプリケーション層）

- **internal/usecase**: ビジネスロジック層（フラットなファイル配置）
  - ビジネスフロー、トランザクション管理を担当
  - 複数のRepositoryを組み合わせた処理

### Domain/Infrastructure層（統合）

- **internal/query**: sqlcで生成されるコード（旧 `repository/sqlc`）
  - `queries/`: SQL クエリファイル（sqlc で処理）
  - 単一のSQLクエリを実行する責務のみ
  - 手動編集禁止
- **internal/model**: ドメインモデル
  - ページに依存しない汎用的なドメインエンティティ（`Work`, `Cast`, `Staff`など）
  - Presentation層に依存しない（`image.Helper`などに依存しない）
- **internal/repository**: Repository層
  - Query結果をModelに変換する
  - 複数のクエリを組み合わせてModelを構築
  - **ModelとRepositoryは1:1の関係**（例: `model.Work` ↔ `repository.WorkRepository`）

### その他

- **cmd/server/main.go**: エントリポイント。設定、データベース接続、Chi ルーターを使用した HTTP サーバーを初期化
- **internal/config**: 環境変数から設定を読み込む設定管理。`.env.{environment}` ファイルを使用
- **internal/auth**: 認証ロジック
- **internal/turnstile**: Cloudflare Turnstile連携

## レイヤー間の依存関係

### 基本方針

```
Presentation層 → Application層 → Domain/Infrastructure層
```

下位層は上位層に依存しません（依存の方向は一方通行）。

### Presentation層（Handler, ViewModel, Template, Middleware）

各パッケージの依存関係：

- **Templates**: `ViewModel` を通じてデータを表示。データアクセス（`repository`, `query`）、ビジネスロジック（`usecase`）、`Model` への直接依存は禁止。
- **ViewModel**: `Model` → `ViewModel` の変換のみ。`repository`, `query` に依存しない
- **Handler**: `query` への直接アクセス禁止。データアクセスは `repository` を経由
- **Middleware**: 共通処理のみ。`query`, `repository`, `usecase`、他のPresentation層パッケージに依存しない

**依存関係の図解**:

```
Templates → ViewModel (OK: 表示用データを受け取る)
              ↓
ViewModel → Model (OK: ドメインデータを表示用に変換)
              ↓
Handler → UseCase, Repository, ViewModel
  ↑
Middleware (独立、他のPresentation層パッケージに依存しない)
```

**重要**: Templates は ViewModel に依存できますが、Model に直接依存することは禁止です。必ず ViewModel を経由してください。

### Application層（UseCase）

- `query` への直接アクセス禁止。データアクセスは `repository` を経由
- Presentation層（`handler`, `middleware`, `viewmodel`, `templates`）に依存しない

### Domain/Infrastructure層（Query, Repository, Model）

各パッケージの依存関係：

- **Model**: 純粋なドメインエンティティ。`query`, `repository` に依存しない
- **Repository**: `query`, `model` に依存できる。上位層に依存しない
- **Query**: sqlc生成コード。他のすべての層に依存しない（独立）

**依存関係の図解**:

```
Query (独立、単独で動作)
  ↓
Repository → Query, Model
  ↓
Model (独立、他に依存しない)
```

**重要**: Repository が Query の結果を Model に変換します。

### 重要なルール

1. **Queryへの依存はRepositoryのみ**: Handler/UseCaseがQueryに直接依存することは禁止
2. **すべてのデータアクセスはRepositoryを経由**: データ取得はRepositoryまたはUseCaseを使う
3. **下位層は上位層に依存しない**: Domain/Infrastructure層はPresentation層に依存しない
4. **関心の分離**: 各パッケージは明確な責務を持ち、その責務に集中する

### なぜRepositoryのみがQueryに依存すべきか

**メリット**:

- ✅ **保守性**: データアクセスロジックがRepositoryに集約される
- ✅ **拡張性**: キャッシュ層の追加、データソース変更がRepositoryのみで完結
- ✅ **一貫性**: 「データ取得 = Repositoryを使う」というルールが明確
- ✅ **テスト容易性**: Repositoryをモックすれば、Handler/UseCaseのテストが容易

**デメリットを回避**:

- ❌ データアクセスロジックの散在（Handler/UseCaseに直接Queryを書く）
- ❌ 変更の波及（データアクセス方法の変更がHandler/UseCaseに影響）
- ❌ ルールの曖昧さ（「このケースはQueryを直接使って良い？」という混乱）

## 物理的な構造と論理的な構造

- **物理的な構造**: `internal/`配下はフラット（機能別にパッケージを分ける）
- **論理的な構造**: ドキュメントでレイヤーごとにパッケージを分類し、依存関係を明示

## ビューモデル（View Model）

### 概要

プレゼンテーション層でのデータ変換は `internal/viewmodel` パッケージで行います。

### 責務

- リポジトリ層のデータ構造をテンプレート表示用の構造に変換
- 画像URL生成、日付フォーマットなどの表示ロジック
- 複数のリポジトリ結果を組み合わせた表示用データの作成

### 命名規則

- **構造体名**: `Work`, `User` など（エンティティ名と同じ）
- **変換関数**: `NewWorkFromXXX` （XXX は sqlc が生成した型名）
- **複数変換**: `NewWorksFromXXX` （複数形）

### 利点

- ハンドラーをシンプルに保つ
- データ変換ロジックの再利用が可能
- テストしやすい構造
- sqlc が生成する型とプレゼンテーション層の分離

### 実装例

```go
// internal/viewmodel/work.go
package viewmodel

import (
    "github.com/annict/annict/internal/repository"
)

// Work はテンプレートで表示する作品データ
type Work struct {
    ID            int64
    Title         string
    ImageURL      string
    WatchersCount int32
    SeasonYear    *int32
    SeasonName    *string
}

// NewWorkFromPopularRow は人気作品クエリの結果をViewModelに変換
func NewWorkFromPopularRow(cfg *config.Config, work repository.GetPopularWorksRow) Work {
    return Work{
        ID:            work.ID,
        Title:         work.Title,
        ImageURL:      generateImageURL(cfg, work.ImageData),
        WatchersCount: work.WatchersCount,
        SeasonYear:    work.SeasonYear,
        SeasonName:    work.SeasonName,
    }
}

// NewWorksFromPopularRows は複数の作品を一括変換
func NewWorksFromPopularRows(cfg *config.Config, works []repository.GetPopularWorksRow) []Work {
    result := make([]Work, len(works))
    for i, work := range works {
        result[i] = NewWorkFromPopularRow(cfg, work)
    }
    return result
}

// generateImageURL は画像データからimgproxy経由のURLを生成
func generateImageURL(cfg *config.Config, imageData *string) string {
    if imageData == nil {
        return ""
    }
    // imgproxy URLを生成
    // ...
}
```

### ハンドラーでの使用

```go
// internal/handler/popular_works.go
package handler

import (
    "github.com/annict/annict/internal/templates/layouts"
    "github.com/annict/annict/internal/templates/pages/works"
    "github.com/annict/annict/internal/viewmodel"
)

func (h *Handler) PopularWorks(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // リポジトリから作品データを取得
    worksRows, err := h.queries.GetPopularWorks(ctx)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // ViewModelに変換
    worksView := viewmodel.NewWorksFromPopularRows(h.cfg, worksRows)

    // ページメタデータを作成
    meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
    meta.SetTitle(ctx, "popular_anime_title")

    // テンプレートをレンダリング
    user := authMiddleware.GetUserFromContext(ctx)
    layouts.Default(ctx, meta, user, works.Popular(ctx, worksView)).Render(ctx, w)
}
```

### テスト

```go
func TestNewWorkFromPopularRow(t *testing.T) {
    cfg := &config.Config{
        Domain: "example.com",
    }

    work := repository.GetPopularWorksRow{
        ID:            1,
        Title:         "テストアニメ",
        WatchersCount: 100,
        ImageData:     stringPtr(`{"id":"test.jpg"}`),
    }

    result := viewmodel.NewWorkFromPopularRow(cfg, work)

    if result.ID != 1 {
        t.Errorf("ID = %d, want 1", result.ID)
    }
    if result.Title != "テストアニメ" {
        t.Errorf("Title = %q, want %q", result.Title, "テストアニメ")
    }
    if result.WatchersCount != 100 {
        t.Errorf("WatchersCount = %d, want 100", result.WatchersCount)
    }
    if result.ImageURL == "" {
        t.Error("ImageURL should not be empty")
    }
}
```

## ユースケース（Use Case）

### 概要

ビジネスロジックとトランザクション管理は `internal/usecase` パッケージで行います。

### 責務

- トランザクション管理（`db.BeginTx` から `tx.Commit` まで）
- 複数の repository を跨ぐ処理
- ビジネスロジックの実装

### ファイル配置

`internal/usecase/` 直下にフラットに配置（サブディレクトリは作成しない）

### 命名規則

- **ファイル名**: `{action}_{entity}.go`
  - 例: `create_session.go`, `create_password_reset_token.go`, `update_password_reset.go`
  - **重要**: 動詞（アクション）を必ず先頭に配置する
- **構造体名**: `{Action}{Entity}Usecase`
  - 例: `CreateSessionUsecase`, `CreatePasswordResetTokenUsecase`
  - 注: `Usecase` の `c` は小文字（既存コードとの統一のため）
- **コンストラクタ**: `New{Action}{Entity}Usecase`
- **Execute メソッド**: `Execute(ctx context.Context, ...) (*Result, error)`

### 結果型

各 UseCase は専用の Result 構造体を返します。

例: `SessionResult`, `CreatePasswordResetTokenResult`

### 利点

- ハンドラーがシンプルになる（HTTP 処理に専念できる）
- トランザクション境界が明確
- テストしやすい構造
- ビジネスロジックの再利用が可能

### 実装例

#### シンプルなユースケース（トランザクションなし）

```go
// internal/usecase/create_session.go
package usecase

import (
    "context"
    "github.com/annict/annict/internal/repository"
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

#### 複雑なユースケース（トランザクションあり）

```go
// internal/usecase/create_password_reset_token.go
package usecase

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    "github.com/annict/annict/internal/repository"
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

### ハンドラーでの使用

```go
// internal/handler/password_reset.go
package handler

import (
    "github.com/annict/annict/internal/usecase"
)

func (h *Handler) ProcessPasswordReset(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // リクエストバリデーション
    req := &PasswordResetRequest{
        Email: r.FormValue("email"),
    }
    if formErrors := req.Validate(ctx); formErrors != nil {
        // エラー処理
        return
    }

    // ユーザーを検索
    user, err := h.queries.GetUserByEmail(ctx, req.Email)
    if err != nil {
        // ユーザーが見つからない場合の処理
        return
    }

    // ユースケースを実行
    uc := usecase.NewCreatePasswordResetTokenUsecase(h.db, h.queries)
    result, err := uc.Execute(ctx, user.ID)
    if err != nil {
        slog.ErrorContext(ctx, "トークンの作成に失敗", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // メール送信
    err = h.sendPasswordResetEmail(ctx, user.Email, result.Token)
    if err != nil {
        slog.ErrorContext(ctx, "メール送信に失敗", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // 成功レスポンス
    http.Redirect(w, r, "/password/reset_sent", http.StatusSeeOther)
}
```

### テスト

```go
func TestCreatePasswordResetTokenUsecase_Execute(t *testing.T) {
    // テストDBとトランザクションをセットアップ
    db, tx := testutil.SetupTestDB(t)
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

### 命名の注意点

#### ファイル名の順序

`{action}_{entity}` の順（動詞を必ず先頭に）

- ✅ `create_session.go`
- ❌ `session_create.go`
- ✅ `create_password_reset_token.go`
- ❌ `password_reset_create_token.go`

#### 複合エンティティ

エンティティが複数単語の場合はアンダースコアで連結

- ✅ `create_password_reset_token.go` （password_reset_token というエンティティ）

#### 構造体名の大文字化

`Usecase` の `c` は小文字

- ✅ `CreateSessionUsecase`
- ❌ `CreateSessionUseCase`

### ファイル配置の理由

#### フラット構造

エンティティごとにディレクトリを作らず、`internal/usecase/` 直下にすべてのファイルを配置

理由:
- **検索性**: ファイル名のプレフィックスでグルーピングされるため、エディタで検索しやすい
- **シンプルさ**: ディレクトリ階層が深くならず、import パスがシンプル
- **スケーラビリティ**: ファイル数が増えても管理しやすい

## ベストプラクティス

### 1. ViewModelとUseCaseの使い分け

```go
// ❌ Bad: ハンドラーで複雑な変換ロジック
func (h *Handler) PopularWorks(w http.ResponseWriter, r *http.Request) {
    works, _ := h.queries.GetPopularWorks(ctx)

    // ハンドラーで複雑な変換を行う（悪い例）
    worksView := make([]WorkView, len(works))
    for i, work := range works {
        imageURL := ""
        if work.ImageData != nil {
            // 複雑な画像URL生成ロジック
            imageURL = generateComplexImageURL(work.ImageData)
        }
        worksView[i] = WorkView{
            ID:       work.ID,
            Title:    work.Title,
            ImageURL: imageURL,
        }
    }
    // ...
}

// ✅ Good: ViewModelで変換
func (h *Handler) PopularWorks(w http.ResponseWriter, r *http.Request) {
    works, _ := h.queries.GetPopularWorks(ctx)
    worksView := viewmodel.NewWorksFromPopularRows(h.cfg, works)
    // ...
}
```

### 2. トランザクション管理はUseCaseで

```go
// ❌ Bad: ハンドラーでトランザクション管理
func (h *Handler) CreateWork(w http.ResponseWriter, r *http.Request) {
    tx, _ := h.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    // 複雑なビジネスロジック
    // ...

    tx.Commit()
}

// ✅ Good: UseCaseでトランザクション管理
func (h *Handler) CreateWork(w http.ResponseWriter, r *http.Request) {
    uc := usecase.NewCreateWorkUsecase(h.db, h.queries)
    result, err := uc.Execute(ctx, params)
    // ...
}
```

### 3. ハンドラーはHTTP処理に専念

```go
// ✅ Good: ハンドラーの責務は明確
func (h *Handler) ProcessPasswordReset(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. リクエストバリデーション（Request DTO）
    req := &PasswordResetRequest{Email: r.FormValue("email")}
    if formErrors := req.Validate(ctx); formErrors != nil {
        // エラーレスポンス
        return
    }

    // 2. ビジネスロジック（UseCase）
    uc := usecase.NewCreatePasswordResetTokenUsecase(h.db, h.queries)
    result, err := uc.Execute(ctx, userID)
    if err != nil {
        // エラーレスポンス
        return
    }

    // 3. レスポンス
    http.Redirect(w, r, "/password/reset_sent", http.StatusSeeOther)
}
```

## まとめ

- **ViewModel**: リポジトリ層のデータをテンプレート表示用に変換
- **UseCase**: ビジネスロジックとトランザクション管理
- **Handler**: HTTP処理に専念し、ViewModelとUseCaseを活用してシンプルに保つ

この構造により、コードの見通しが良く、テストしやすく、保守しやすいアーキテクチャを実現できます。
