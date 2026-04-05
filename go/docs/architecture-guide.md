# アーキテクチャガイド

このドキュメントは、Go版Annictのアーキテクチャパターンを説明します。

## 概要

Go版Annictは、関心の分離を意識した**3層アーキテクチャ**を採用しています。

### 3層アーキテクチャの構成

```
┌─────────────────────────────────────────────────────────┐
│ Presentation層（プレゼンテーション層）                    │
│ - Handler, Worker                                      │
│ - ViewModel                                            │
│ - Template                                             │
│ - Middleware                                           │
│ - Presentation層のヘルパー（image, i18n, session）       │
└─────────────────────────────────────────────────────────┘
         ↓ 依存（OK）
┌─────────────────────────────────────────────────────────┐
│ Application層（アプリケーション層）                        │
│ - UseCase, Validator                                   │
└─────────────────────────────────────────────────────────┘
         ↓ 依存（OK）
┌─────────────────────────────────────────────────────────┐
│ Domain/Infrastructure層（統合）                          │
│ - Query (sqlc)                                         │
│ - Repository                                           │
│ - Model                                                │
│ - Dispatcher                                           │
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
- `model.UserCalendar` ↔ `repository.UserCalendarRepository`

この1:1関係により、以下のメリットがあります：

- **一貫性**: どのModelに対してどのRepositoryを使うかが明確
- **保守性**: Modelの変更がRepositoryに集約される
- **可読性**: コードの見通しが良くなる

#### 命名規則

Model と Repository のファイル名・構造体名は統一します：

| Model          | Repository               | ファイル名         |
| -------------- | ------------------------ | ------------------ |
| `Work`         | `WorkRepository`         | `work.go`          |
| `User`         | `UserRepository`         | `user.go`          |
| `UserCalendar` | `UserCalendarRepository` | `user_calendar.go` |

**命名のルール**:

- **ファイル名**: スネークケース（`user_calendar.go`）
- **構造体名**: パスカルケース（`UserCalendar`, `UserCalendarRepository`）
- **Model と Repository は同じ名前**: `model/user_calendar.go` ↔ `repository/user_calendar.go`

#### モデルの重複を避ける

クエリの結果や状態ごとに新しいモデルを作らず、既存のモデルを再利用します。関連エンティティのデータが必要な場合は、ポインタ型のフィールドでモデル間の参照を表現します。

```go
// ✅ 良い例: 既存の Work モデルに User への参照を持たせる
type Work struct {
    ID    WorkID
    User  *User  // 関連エンティティへのポインタ参照
    Title string
}

// ❌ 悪い例: クエリ結果に合わせた専用モデルを作る
type JoinedWork struct {
    WorkID    WorkID
    WorkTitle string
    UserID    UserID
    UserName  string
}
```

Repositoryではクエリ結果ごとに変換メソッドを用意し、同じモデルに変換します：

```go
// 単純なクエリ結果 → Work（User は ID のみ）
func (r *WorkRepository) toModel(row query.Work) *model.Work { ... }

// JOINクエリ結果 → Work（User のフィールドをより多く設定）
func (r *WorkRepository) toWorksFromJoinedRows(rows []query.ListJoinedWorksByUserRow) []*model.Work { ... }
```

#### Queryファイルの命名

Queryファイルは用途に応じて2つのパターンがあります：

**1. テーブル名ベース（単純なCRUD操作）**:

- 単一テーブルに対するCRUD操作
- 例: `users.sql`, `works.sql`, `sessions.sql`

**2. モデル/機能名ベース（複雑なクエリ）**:

- 複数テーブルをJOINするクエリ
- 特定のモデルを構築するためのクエリ
- 例: `user_calendar.sql`（users, library_entries, slots, works をJOIN）

```
internal/query/queries/
├── users.sql           # usersテーブルのCRUD
├── works.sql           # worksテーブルのCRUD
├── sessions.sql        # sessionsテーブルのCRUD
└── user_calendar.sql   # UserCalendarモデル用の複合クエリ
```

### データの流れ

1. **Query** (Domain/Infrastructure層): SQLクエリを実行し、クエリ結果（`query.GetPopularWorksRow`など）を返す
2. **Repository** (Domain/Infrastructure層): Query結果をModelに変換し、複数のクエリを組み合わせる
3. **Model** (Domain/Infrastructure層): ページに依存しない汎用的なドメインエンティティ（`model.Work`など）
4. **UseCase** (Application層): Repository経由でModelを取得し、ビジネスロジックを実行して結果を返す
5. **Handler** (Presentation層): UseCaseを呼び出してModelを取得し、ModelをViewModelに変換
6. **ViewModel** (Presentation層): 表示用のデータ構造（画像URL生成、言語切り替えなど）
7. **Template** (Presentation層): ViewModelを受け取ってHTMLを生成

### 主要なレイヤー

- **Handler**: HTTPリクエスト・レスポンスの処理（薄いAdapter）
- **Worker**: バックグラウンドジョブの受信・処理（薄いAdapter）
- **UseCase**: ビジネスロジック・バリデーション・データ取得のオーケストレ���ション
- **Validator**: 入力値の形式チェックとDBを使った検証
- **ViewModel**: プレゼンテーション層のデータ変換
- **Repository**: Query結果をModelに変換
- **Model**: ドメインエンティティ
- **Query**: sqlc生成コード（データアクセス層）
- **Dispatcher**: ジョブキューへの投入

## レイヤーごとのパッケージ分類

### Presentation層（プレゼンテーション層）

- **internal/handler**: HTTP リクエストハンドラー（リソースごとにディレクトリを切り、1 エンドポイント = 1 ファイルの原則）
  - 薄い Adapter として HTTP の入出力変換のみを担当
  - すべてのデータアクセスは UseCase を経由する（Repository への直接依存は禁止）
- **internal/worker**: バックグラウンドジョブワーカー
  - 薄い Adapter としてジョブの受信・UseCase 呼び出しのみを担当
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
  - Handler のすべてのデータアクセスのオーケストレーター
  - 書き込み UseCase: バリデーション（Validator 呼び出し）、ビジネスロジック、トランザクション管理
  - 読み取り UseCase: Repository 経由でデータを取得して返却
- **internal/validator**: バリデーション（フラットなファイル配置）
  - 形式バリデーション（入力値の形式チェック）と状態バリデーション（DB を使った検証）を統合
  - UseCase から呼び出される（Handler から直接呼び出さない）

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
- **internal/dispatcher**: ジョブキューへの投入
  - バックグラウンドジョブの投入を担当（UseCase から呼び出される）
  - UseCase が `worker` パッケージに直接依存することを防ぐ

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

### Presentation層（Handler, Worker, ViewModel, Template, Middleware）

各パッケージの依存関係：

- **Templates**: `ViewModel` を通じてデータを表示。データアクセス（`repository`, `query`）、ビジネスロジック（`usecase`）、`Model` への直接依存は禁止。
- **ViewModel**: `Model` → `ViewModel` の変換のみ。`repository`, `query` に依存しない
- **Handler**: `query`, `repository`, `validator` への直接アクセス禁止。すべてのデータアクセスは `usecase` を経由
- **Worker**: 薄い Adapter。`query`, `repository` への直接アクセス禁止。ジョブの��信と `usecase` 呼び出しのみ
- **Middleware**: 共通処理のみ。`query`, `repository`, `usecase`、`handler`, `viewmodel` に依存しない。エラーページやメンテナンスページのレンダリングのため `templates` への依存は許可

**依存関係の図解**:

```
Templates → ViewModel (OK: 表示用データを受け取る)
              ↓
ViewModel → Model (OK: ドメインデータを表示用に変換)
              ↓
Handler → UseCase, ViewModel (Repository・Validator への直接依存は禁止)
Worker → UseCase (Repository・Validator への直接依存は禁止)
  ↑
Middleware → Templates (OK: エラーページ等のレンダリング)
```

**重要**: Templates は ViewModel に依存できますが、Model に直接依存することは禁止です。必ず ViewModel を経由してください。

### Application層（UseCase, Validator）

- **UseCase**: すべてのデータアクセスのオーケストレーター。`query` への直接アクセス禁��。データアクセスは `repository` を経由。書き込み UseCase は `validator` を呼び出してバリデーションを実行
- **Validator**: 入力値の形式チェックと DB を使った検証。`repository` に依存可能。Presentation層（`handler`, `middleware`, `viewmodel`, `templates`）や `usecase` に依存しない
- Presentation層（`handler`, `worker`, `middleware`, `viewmodel`, `templates`）に依存しない

### Domain/Infrastructure層（Query, Repository, Model, Dispatcher）

各パッケージの依存関係：

- **Model**: 純粋なドメインエンティティ。`query`, `repository` に依存しない
- **Repository**: `query`, `model` に依存できる。上位層に依存しない
- **Query**: sqlc生成コード。他のすべての層に依存しない（独立）
- **Dispatcher**: ジョブキューへの投入。`worker` パッケージのジョブ引数型を import する。上位層に依存しない

**依存関係の図解**:

```
Query (独立、単独で動作)
  ↓
Repository → Query, Model
  ↓
Model (独立、他に依存しない)

Dispatcher → worker（ジョブ引数型のみ）
```

**重要**: Repository が Query の結果を Model に変換します。

### 重要なルール

1. **Handler → Repository の直接依存は禁止**: Handler のすべてのデータアクセスは UseCase を経由する
2. **Queryへの依存はRepositoryのみ**: Handler/UseCaseがQueryに直接依存することは禁止
3. **Validator は UseCase から呼び出す**: Handler から直接 Validator を呼び出さない
4. **下位層は上位層に依存しない**: Domain/Infrastructure層はPresentation層に依存しない
5. **関心の分離**: 各パッケージは明確な責務を持ち、その責務に集中する

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
    "github.com/annict/annict/go/internal/repository"
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
    "github.com/annict/annict/go/internal/templates/layouts"
    "github.com/annict/annict/go/internal/templates/pages/works"
    "github.com/annict/annict/go/internal/viewmodel"
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

Handler のすべてのデータアクセスのオーケストレーターとして `internal/usecase` パッケージを使用します。読み取り・書き込みの両方を UseCase で担当します。

### UseCase の種類

**書き込み UseCase**:

- バリデーション（Validator 呼び出し）、ビジネスロジック、トランザクション管理を統括
- 複数のRepositoryを跨ぐ永続化処理
- ロールバックが必要な複合操作

```go
// 例: バリデーション + 永続化を統括する書き込み UseCase
type SendSignInCodeUsecase struct {
    validator *validator.CreateSignInValidator
    userRepo  *repository.UserRepository
    // ...
}

func (uc *SendSignInCodeUsecase) Execute(ctx context.Context, input SendSignInCodeInput) (*SendSignInCodeOutput, error) {
    // 1. バリデーション
    valResult := uc.validator.Validate(ctx, validator.CreateSignInValidatorInput{
        Email: input.Email,
    })
    if valResult.FormErrors.HasErrors() {
        return &SendSignInCodeOutput{FormErrors: valResult.FormErrors}, nil
    }

    // 2. ビジネスロジック + 永続化
    user, err := uc.userRepo.GetByEmailForSignIn(ctx, input.Email)
    // ...
}
```

**読み取り UseCase**:

- Repository 経由でデータを取得して返却
- トランザクション不要

```go
// 例: 人気作品の取得（読み取り UseCase）
type GetPopularWorksUsecase struct {
    workRepo *repository.WorkRepository
}

func (uc *GetPopularWorksUsecase) Execute(ctx context.Context) (*GetPopularWorksOutput, error) {
    works, err := uc.workRepo.GetPopularWorksWithDetails(ctx)
    if err != nil {
        return nil, fmt.Errorf("人気作品の取得に失敗: %w", err)
    }
    return &GetPopularWorksOutput{Works: works}, nil
}
```

**判断基準**: Handler から Repository を直接呼び出さない。読み取り専用の処理にも UseCase を作成し、Handler → UseCase → Repository の流れに統一する。

### 責務

- Handler のすべてのデータアクセスのオーケストレーション
- バリデーション（Validator 呼び出し）
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
    "github.com/annict/annict/go/internal/repository"
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
    "github.com/annict/annict/go/internal/repository"
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
    "github.com/annict/annict/go/internal/usecase"
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

### Repository の WithTx パターン

Usecase でトランザクションを使用する場合、**Repository の `WithTx` メソッド**を使用してトランザクション内で操作するリポジトリを取得します。

#### なぜ WithTx パターンを使うのか

**メリット**:

- **依存性注入**: Repository をコンストラクタで受け取るため、テストでモックに差し替えやすい
- **意図が明確**: `WithTx(tx)` の呼び出しで「このリポジトリはトランザクション内で操作する」という意図が明確
- **一貫性**: すべての Usecase で同じパターンを使用することで、コードの読みやすさが向上

#### Repository に WithTx を実装する

各 Repository には `WithTx` メソッドを実装します：

```go
// internal/repository/user_repository.go

// WithTx はトランザクションを使用する新しいRepositoryを返す
func (r *UserRepository) WithTx(tx *sql.Tx) *UserRepository {
    return &UserRepository{q: r.q.WithTx(tx)}
}
```

#### Usecase で WithTx を使用する

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

#### 重要なポイント

1. **Repository はコンストラクタで受け取る**: `NewXxxUsecase` で Repository を引数として受け取る
2. **Execute 内で WithTx を呼び出す**: トランザクションを開始した後、各 Repository の `WithTx(tx)` を呼び出す
3. **元の Repository は変更しない**: `WithTx` は新しい Repository インスタンスを返すため、元の Repository には影響しない
4. **すべての Repository で WithTx を使う**: トランザクション内で使用するすべての Repository に対して `WithTx` を呼び出す

### テスト

```go
func TestCreatePasswordResetTokenUsecase_Execute(t *testing.T) {
    // テストDBとトランザクションをセットアップ
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
