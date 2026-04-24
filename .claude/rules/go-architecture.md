---
paths:
  - "go/**/*.{go,templ}"
---

# アーキテクチャガイド

このドキュメントは、Go版Wikinoのアーキテクチャパターンを説明します。

## 概要

Go版Wikinoは、関心の分離を意識した**3層アーキテクチャ**を採用しています。

### 3層アーキテクチャの構成

```
┌─────────────────────────────────────────────────────────┐
│ Presentation層（プレゼンテーション層）                    │
│ - Handler（HTTP Adapter）                              │
│ - Worker（Job Adapter）                                │
│ - ViewModel                                            │
│ - Template                                             │
│ - Middleware                                           │
│ - Presentation層のヘルパー（image, i18n, session）       │
└─────────────────────────────────────────────────────────┘
         ↓ 依存（OK）
┌─────────────────────────────────────────────────────────┐
│ Application層（アプリケーション層）                        │
│ - UseCase（オーケストレーター：認可・検証・ロジック・永続化）│
│ - Validator（入力バリデーション）                         │
└─────────────────────────────────────────────────────────┘
         ↓ 依存（OK）
┌─────────────────────────────────────────────────────────┐
│ Domain/Infrastructure層（統合）                          │
│ - Query (sqlc)                                         │
│ - Repository                                           │
│ - Model                                                │
│ - Dispatcher（ジョブキュー投入）                         │
│ （同じ層なので相互に依存できる）                          │
└─────────────────────────────────────────────────────────┘
```

**重要**: Domain/Infrastructure層はPresentation層に**依存してはいけない**

### Domain/Infrastructure層を統合する理由

このプロジェクトでは、Domain層とInfrastructure層を分離せず、統合して扱います：

- **実用的**: データベース変更（PostgreSQL → MySQLなど）はほぼ起こらない
- **シンプル**: 層をまたぐ変換コストを削減し、シンプルさを保つ
- **Goらしい**: Goのプラグマティックな哲学に合致

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

#### 命名規則

ModelとRepositoryのファイル名・構造体名は統一します：

| Model         | Repository              | ファイル名        |
| ------------- | ----------------------- | ----------------- |
| `Page`        | `PageRepository`        | `page.go`         |
| `User`        | `UserRepository`        | `user.go`         |
| `SpaceMember` | `SpaceMemberRepository` | `space_member.go` |

**命名のルール**:

- **ファイル名**: スネークケース（`space_member.go`）
- **構造体名**: パスカルケース（`SpaceMember`, `SpaceMemberRepository`）
- **ModelとRepositoryは同じ名前**: `model/space_member.go` ↔ `repository/space_member.go`

#### モデルの重複を避ける

クエリの結果や状態ごとに新しいモデルを作らず、既存のモデルを再利用します。関連エンティティのデータが必要な場合は、ポインタ型のフィールドでモデル間の参照を表現します。

```go
// ✅ 良い例: 既存の Topic モデルに Space への参照を持たせる
type Topic struct {
    ID     TopicID
    Space  *Space  // 関連エンティティへのポインタ参照
    Number int32
    Name   string
}

// ❌ 悪い例: クエリ結果に合わせた専用モデルを作る
type JoinedTopic struct {
    TopicID         TopicID
    TopicName       string
    SpaceID         SpaceID
    SpaceIdentifier SpaceIdentifier
    SpaceName       string
}
```

Repositoryではクエリ結果ごとに変換メソッドを用意し、同じモデルに変換します：

```go
// 単純なクエリ結果 → Topic（Space は ID のみ）
func (r *TopicRepository) toModel(row query.Topic) *model.Topic { ... }

// JOINクエリ結果 → Topic（Space のフィールドをより多く設定）
func (r *TopicRepository) toTopicsFromJoinedRows(rows []query.ListJoinedTopicsByUserRow) []*model.Topic { ... }
```

#### ドメインID型

モデルのIDフィールドには `string` ではなく、`internal/model/id.go` に定義された専用のドメインID型（`type SpaceID string` 等）を使用します。これにより、異なるエンティティのIDを取り違える問題をコンパイル時に検出できます。

**基本ルール**:

- ✅ **モデルのIDフィールドには専用型を使用**: `ID SpaceID`、`ID PageID` など
- ✅ **外部キーにも専用型を使用**: `UserID model.UserID`、`SpaceID model.SpaceID` など
- ✅ **新しいモデルを追加する場合は対応するID型も追加**: `id.go` に型と `String()` メソッドを定義
- ❌ **IDフィールドに `string` を使用しない**

**実装パターン**:

```go
// internal/model/id.go - 型定義
type SpaceID string
func (id SpaceID) String() string { return string(id) }

// internal/model/space.go - モデルでの使用
type Space struct {
    ID   SpaceID
    Name string
}

// internal/repository/space.go - リポジトリでの変換（sqlcのstringから専用型へ）
func toModel(row query.GetSpaceRow) *model.Space {
    return &model.Space{
        ID:   model.SpaceID(row.ID),
        Name: row.Name,
    }
}

// internal/testutil/space_builder.go - テストビルダーの戻り値
func (b *SpaceBuilder) Build() model.SpaceID {
    // ...
    return model.SpaceID(id)
}
```

**スライス変換ヘルパー**:

IDのスライスと `[]string` の相互変換が必要な場合（例: PostgreSQL の配列型との変換）は、`id.go` にヘルパー関数を定義します：

```go
func PageIDsToStrings(ids []PageID) []string { ... }
func StringsToPageIDs(ss []string) []PageID { ... }
```

#### Queryファイルの命名

Queryファイルは用途に応じて2つのパターンがあります：

**1. テーブル名ベース（単純なCRUD操作）**:

- 単一テーブルに対するCRUD操作
- 例: `users.sql`, `pages.sql`, `sessions.sql`

**2. モデル/機能名ベース（複雑なクエリ）**:

- 複数テーブルをJOINするクエリ
- 特定のモデルを構築するためのクエリ
- 例: `space_member.sql`（users, space_membersなどをJOIN）

```
internal/query/queries/
├── users.sql           # usersテーブルのCRUD
├── pages.sql           # pagesテーブルのCRUD
├── sessions.sql        # sessionsテーブルのCRUD
└── space_member.sql    # SpaceMemberモデル用の複合クエリ
```

### データの流れ

1. **Query** (Domain/Infrastructure層): SQLクエリを実行し、クエリ結果（`query.GetPopularWorksRow`など）を返す
2. **Repository** (Domain/Infrastructure層): Query結果をModelに変換し、複数のクエリを組み合わせる
3. **Model** (Domain/Infrastructure層): ページに依存しない汎用的なドメインエンティティ（`model.Work`など）
4. **Handler** (Presentation層): UseCaseからModelを取得し、ModelをViewModelに変換
5. **ViewModel** (Presentation層): 表示用のデータ構造（画像URL生成、言語切り替えなど）
6. **Template** (Presentation層): ViewModelを受け取ってHTMLを生成

### 主要なレイヤー

- **Handler**: HTTP リクエストのパース → UseCase 呼び出し → レスポンス生成（薄い Adapter）
- **Worker**: ジョブ引数の変換 → UseCase 呼び出し（薄い Adapter）
- **UseCase**: オーケストレーター（データ取得・認可・バリデーション・ビジネスロジック・永続化を統括）
- **ViewModel**: プレゼンテーション層のデータ変換
- **Validator**: バリデーション（`internal/validator/` に全て配置、UseCase から呼び出される）
- **Policy**: 認可ロジック（Model のみに依存、UseCase から呼び出される）
- **Repository**: Query結果をModelに変換
- **Dispatcher**: ジョブキューへの投入を抽象化
- **Model**: ドメインエンティティ
- **Query**: sqlc生成コード（データアクセス層）

## レイヤーごとのパッケージ分類

### Presentation層（プレゼンテーション層）

- **internal/handler**: HTTP リクエストハンドラー（薄い Adapter。リソースごとにディレクトリを切り、1 エンドポイント = 1 ファイルの原則）
  - リクエストのパース → UseCase 呼び出し → レスポンス（リダイレクト or テンプレート描画）
  - UseCase からのエラーを `errors.As` で判別してレスポンスを決定する
- **internal/worker**: バックグラウンドジョブの受信（薄い Adapter）
  - ジョブ引数の変換 → UseCase 呼び出しのみ
  - ビジネスロジックやテンプレートレンダリングは含まない
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

- **internal/email**: メール送信・テンプレートレンダリング（メール種別ごとの Sender でテンプレートレンダリングと i18n 件名取得を担当）
- **internal/image**: 画像URL生成（imgproxy署名付きURL）
- **internal/i18n**: 国際化（翻訳取得、言語切り替え）
- **internal/session**: セッション管理（フラッシュメッセージ、ユーザー情報）

### Application層（アプリケーション層）

- **internal/usecase**: オーケストレーション層（フラットなファイル配置）
  - 読み取り UseCase: データ取得、複数 Repository の集約（トランザクションなし）
  - 書き込み UseCase: 認可チェック → バリデーション → ビジネスロジック → トランザクション管理
  - Handler/Worker からのすべてのデータアクセスは UseCase を経由する
- **internal/validator**: バリデーション（形式チェック + 状態チェック）
  - すべてのバリデーターをこのパッケージに配置（形式バリデーションのみの場合も含む）
  - UseCase から呼び出され、状態バリデーションでは Repository に依存する
  - `main.go` で構築し、UseCase のコンストラクタに渡す

### Domain/Infrastructure層（統合）

- **internal/query**: sqlcで生成されるコード（旧 `repository/sqlc`）
  - `queries/`: SQL クエリファイル（sqlc で処理）
  - 単一のSQLクエリを実行する責務のみ
  - 手動編集禁止
- **internal/model**: ドメインモデル
  - ページに依存しない汎用的なドメインエンティティ（`Work`, `Cast`, `Staff`など）
  - Presentation層に依存しない（`image.Helper`などに依存しない）
  - IDフィールドには `string` ではなく専用のドメインID型（`id.go` に定義）を使用する
- **internal/repository**: Repository層
  - Query結果をModelに変換する
  - 複数のクエリを組み合わせてModelを構築
  - **ModelとRepositoryは1:1の関係**（例: `model.Work` ↔ `repository.WorkRepository`）
- **internal/policy**: 認可ロジック
  - リソースに対する権限判定（例: `TopicPolicy` でページの作成・更新権限を判定）
  - Model のみに依存し、Query・Repository・UseCase には依存しない
  - UseCase から呼び出される
- **internal/dispatcher**: ジョブキューへの投入を抽象化
  - Repository がデータベースアクセスを抽象化するのと同じ発想で、ジョブキューアクセスを抽象化する
  - River（外部ライブラリ）のみに依存。上位層（UseCase, Handler, Worker）には依存しない
  - ジョブ引数の Args 型もこのパッケージ内に定義する

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
- **Handler**: 薄い HTTP Adapter。`query`, `repository`, `validator`, `policy` への直接アクセス禁止。すべてのデータアクセス・認可・バリデーションは `usecase` を経由する。UseCase からのエラーを `errors.As` で判別してレスポンスを決定する
- **Worker**: 薄い Job Adapter。ジョブ引数を UseCase の入力に変換して呼び出すだけ。`templates`, `repository`, `query`, `validator`, `policy` への依存は禁止
- **Middleware**: 共通処理のみ。`query`, `repository`, `usecase`、`handler`, `viewmodel` に依存しない。エラーページやメンテナンスページのレンダリングのため `templates` への依存は許可

**依存関係の図解**:

```
Templates → ViewModel (OK: 表示用データを受け取る)
              ↓
ViewModel → Model (OK: ドメインデータを表示用に変換)
              ↓
Handler → UseCase, ViewModel
Worker  → UseCase, Dispatcher (Args 型の参照)
  ↑
Middleware → Templates (OK: エラーページ等のレンダリング)
```

**重要**: Templates は ViewModel に依存できますが、Model に直接依存することは禁止です。必ず ViewModel を経由してください。

### Application層（UseCase, Validator）

- `query` への直接アクセス禁止。データアクセスは `repository` を経由
- Presentation層（`handler`, `worker`, `middleware`, `viewmodel`, `templates`）に依存しない
- **UseCase → Policy, Validator, Dispatcher**: UseCase はオーケストレーターとして Policy・Validator・Dispatcher に依存する
- **UseCase → email（interface 経由）**: メール送信は UseCase 側で定義した interface に依存し、`templates` を直接 import しない

### Domain/Infrastructure層（Query, Repository, Model, Policy, Dispatcher）

各パッケージの依存関係：

- **Model**: 純粋なドメインエンティティ。`query`, `repository` に依存しない
- **Policy**: 認可ロジック。`model` のみに依存し、`query`, `repository` に依存しない
- **Repository**: `query`, `model` に依存できる。上位層に依存しない
- **Dispatcher**: ジョブキューへの投入。River（外部ライブラリ）のみに依存。上位層（UseCase, Handler, Worker）や同レイヤーの他パッケージ（Validator, Policy, Repository, Model）には依存しない
- **Query**: sqlc生成コード。他のすべての層に依存しない（独立）

**依存関係の図解**:

```
Query (独立、単独で動作)
  ↓
Repository → Query, Model
  ↓
Policy → Model
  ↓
Dispatcher → River (外部ライブラリのみ)
  ↓
Model (独立、他に依存しない)
```

**重要**: Repository が Query の結果を Model に変換します。

### 重要なルール

1. **Queryへの依存はRepositoryのみ**: Handler/UseCaseがQueryに直接依存することは禁止
2. **HandlerからRepositoryへの直接依存は禁止**: HandlerのすべてのデータアクセスはUseCaseを経由する
3. **下位層は上位層に依存しない**: Domain/Infrastructure層はPresentation層に依存しない
4. **関心の分離**: 各パッケージは明確な責務を持ち、その責務に集中する
5. **ValidatorはApplication層**: すべてのバリデーションは `internal/validator/` に配置する。UseCase から呼び出される
6. **認可チェックはUseCaseで実行**: UseCase がオーケストレーターとして Policy を呼び出し認可チェックを行う。Handler から Policy への直接依存は禁止
7. **HandlerからValidator・Policyへの直接依存は禁止**: Handler はすべて UseCase を経由する
8. **WorkerからTemplatesへの依存は禁止**: メールレンダリングは UseCase を経由する
9. **UseCaseからTemplatesへの依存は禁止**: メールレンダリングは email パッケージが担当し、UseCase は interface 経由で呼び出す

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

### 設計方針: 画面の要件に応じて定義する

ViewModelはModelとは異なり、**画面の要件に応じて必要な数だけ定義**します。

**ModelとViewModelの設計方針の違い**:

| 層        | 方針                                   | 理由                                 |
| --------- | -------------------------------------- | ------------------------------------ |
| Model     | ドメイン概念と1:1、重複しない          | ドメインの真実を1箇所に集約するため  |
| ViewModel | 画面の要件に応じて必要な数だけ定義する | 画面ごとに必要な表示項目が異なるため |

**理由**:

- **責務が異なる**: Modelはドメインの概念を表現し、ViewModelは「画面に何を表示するか」を表現する。画面ごとに必要な情報が異なるのは自然
- **変更の理由が異なる**: Modelはドメインルールの変更で変わるが、ViewModelはUIの変更で変わる。1つのViewModelを複数画面で共有すると、ある画面のUI変更が他の画面に波及するリスクがある
- **フィールドの肥大化を防ぐ**: 無理に共有すると「このフィールドはどの画面で使うのか」が分かりにくくなる

**ただし**: 表示項目が同じであれば再利用しても構いません。重複を避けること自体が目的ではなく、「画面の要件に合ったViewModelを定義する」のが原則です。

```go
// ✅ 良い例: 画面ごとに異なるViewModelを定義
// サイドバー用（シンプル）
type TopicForSidebar struct {
    Name   string
    Number int32
    Space  SpaceForSidebar
}

// 詳細ページ用（詳細な情報を含む）
type TopicForDetail struct {
    Name        string
    Number      int32
    Description string
    MemberCount int32
    IconName    IconName
}

// ✅ 良い例: 表示項目が同じなら共有しても良い
// 複数の画面で同じ表示をする場合は1つのViewModelを再利用
type Topic struct {
    Name     string
    Number   int32
    IconName IconName
}
```

### Presentation 層用の型定義

Templates は Model に直接依存できない（depguard で禁止）ため、パスヘルパー関数などで型安全性が必要な場合は ViewModel パッケージに Presentation 層用の型を定義する。

```go
// internal/viewmodel/page_number.go
// Model の型をラップした Presentation 層用の型
type PageNumber model.PageNumber
```

Templates のパスヘルパー関数は ViewModel の型を引数に取る：

```go
// internal/templates/path.go
func PagePath(spaceIdentifier string, pageNumber viewmodel.PageNumber) Path { ... }
```

ViewModel のコンストラクタで `model.PageNumber` → `viewmodel.PageNumber` の変換を行う：

```go
// internal/viewmodel/suggestion.go
diffs[i] = SuggestionPageDiff{
    PageNumber: PageNumber(pageNumberByID[sp.PageID]),  // model.PageNumber → viewmodel.PageNumber
}
```

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
    "github.com/wikinoapp/wikino/internal/repository"
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
    "github.com/wikinoapp/wikino/internal/templates/layouts"
    "github.com/wikinoapp/wikino/internal/templates/pages/works"
    "github.com/wikinoapp/wikino/internal/viewmodel"
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

📖 **詳細は [docs/usecase-guide.md](usecase-guide.md) を参照**

## ワーカー（Worker）

バックグラウンドジョブの受信を担当する薄い Adapter です。River（PostgreSQL ベースのジョブキュー）を使用します。Worker は Presentation 層に属し、Handler と同じく UseCase を呼び出すだけの役割です。

- **配置**: `internal/worker`（Presentation 層）
- **責務**: ジョブ引数の変換 → UseCase 呼び出し
- **命名**: ファイル名 `{action}_{entity}.go`、構造体名 `{Action}{Entity}Worker`
- **ジョブ引数**: `internal/dispatcher/` パッケージ内の `{Action}{Entity}Args` 構造体を使用

**Worker の依存関係**:

- Worker は **UseCase を呼び出す**（すべてのビジネスロジックは UseCase に委譲）
- Worker は **Dispatcher の Args 型を参照する**（ジョブ引数の型定義）
- Worker は **templates に依存不可**（メールレンダリングは UseCase を経由）
- Worker は **Repository, Query, Validator, Policy に依存不可**

```go
// ワーカーの実装例（薄い Adapter）
type SendEmailConfirmationWorker struct {
    river.WorkerDefaults[dispatcher.SendEmailConfirmationArgs]
    uc *usecase.SendEmailConfirmationUsecase
}

func (w *SendEmailConfirmationWorker) Work(ctx context.Context, job *river.Job[dispatcher.SendEmailConfirmationArgs]) error {
    return w.uc.Execute(ctx, usecase.SendEmailConfirmationInput{
        Email:  job.Args.Email,
        Code:   job.Args.Code,
        Locale: job.Args.Locale,
    })
}
```

## ディスパッチャー（Dispatcher）

ジョブキューへの投入を抽象化します。Repository がデータベースアクセスを抽象化するのと同じ発想で、Dispatcher がジョブキューアクセスを抽象化します。

- **配置**: `internal/dispatcher`（Domain/Infrastructure 層）
- **責務**: ジョブキューへの投入、Args 型の定義
- **依存先**: River（外部ライブラリ）のみ。上位層（UseCase, Handler, Worker）には依存しない

**Repository との対比**:

|                          | Repository                                       | Dispatcher                                      |
| ------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| **抽象化する対象**       | データベースアクセス（同期的なデータ永続化）     | ジョブキューへの投入（非同期タスク委譲）        |
| **層**                   | Domain/Infrastructure                            | Domain/Infrastructure                           |
| **UseCase からの見え方** | `repo.FindByID(ctx, id)`                         | `dispatcher.EnqueueEmailConfirmation(ctx, ...)` |
| **分離の基準**           | インフラの種類ではなく、**操作の性質**で分離する |

```go
// internal/dispatcher/dispatcher.go
package dispatcher

type Dispatcher struct {
    client JobInserter
}

func NewDispatcher(client JobInserter) *Dispatcher {
    return &Dispatcher{client: client}
}

func (d *Dispatcher) EnqueueEmailConfirmation(ctx context.Context, email, code, locale string) error {
    _, err := d.client.Insert(ctx, &SendEmailConfirmationArgs{
        Email: email, Code: code, Locale: locale,
    }, nil)
    return err
}

// Args 型もこのパッケージ内に定義する
type SendEmailConfirmationArgs struct {
    Email  string `json:"email"`
    Code   string `json:"code"`
    Locale string `json:"locale"`
}
func (SendEmailConfirmationArgs) Kind() string { return "send_email_confirmation" }
```

**依存の方向**:

```
Worker (Presentation)      → dispatcher (Args 型参照) + usecase
UseCase (Application)      → dispatcher (Enqueue メソッド呼び出し)
Dispatcher (Domain/Infra)  → river（外部ライブラリ）
```

循環なし。UseCase は River やジョブ Args 型の存在を知らない。

## メール送信（Email）

メール送信とテンプレートレンダリングを担当する Presentation 層のヘルパーパッケージです。メール種別ごとの Sender を提供し、テンプレートレンダリングと i18n 件名取得を email パッケージ内に閉じ込めます。

- **配置**: `internal/email`（Presentation 層ヘルパー）
- **責務**: メール送信、テンプレートレンダリング、i18n 件名取得
- **依存先**: `templates`（メールテンプレート）、`i18n`（件名の翻訳）、外部ライブラリ（`templ`, `resend`）

### パッケージ構成

| ファイル                   | 責務                                                    |
| -------------------------- | ------------------------------------------------------- |
| `sender.go`                | `Sender` インターフェース、`ResendSender`、`NoopSender` |
| `confirmation_sender.go`   | メール確認コードのテンプレートレンダリング + 送信       |
| `password_reset_sender.go` | パスワードリセットのテンプレートレンダリング + 送信     |

### Sender インターフェース

`Sender` はメール送信の基盤インターフェースです。`templ.Component` をレンダリングして送信します。

```go
// internal/email/sender.go
type Sender interface {
    Send(ctx context.Context, input SendInput) error
}

type SendInput struct {
    To       string          // 送信先メールアドレス
    Subject  string          // 件名
    HTMLBody templ.Component // メール本文（HTML形式）
    TextBody templ.Component // メール本文（テキスト形式）
}
```

実装:

- `ResendSender`: Resend API を使用した本番用の送信実装
- `NoopSender`: メールを送信せず記録のみ行うテスト用実装

### メール種別ごとの Sender

メール種別ごとに専用の Sender を定義し、テンプレートレンダリングと i18n 件名取得の責務を担います。

```go
// internal/email/confirmation_sender.go
type ConfirmationSender struct {
    sender Sender
}

func (s *ConfirmationSender) Send(ctx context.Context, to, code, appURL, locale string) error {
    ctx = i18n.SetLocale(ctx, locale)
    subject := i18n.T(ctx, "email_confirmation_subject")

    // テンプレートを選択してレンダリング
    var htmlBody, textBody templ.Component
    switch locale {
    case "ja":
        htmlBody = email_confirmation.JaHTML(data)
        textBody = email_confirmation.JaText(data)
    default:
        htmlBody = email_confirmation.EnHTML(data)
        textBody = email_confirmation.EnText(data)
    }

    return s.sender.Send(ctx, SendInput{To: to, Subject: subject, HTMLBody: htmlBody, TextBody: textBody})
}
```

### UseCase との連携（interface パターン）

UseCase は email パッケージに直接依存せず、UseCase 側で定義した小さな interface に依存します。これにより UseCase は `internal/templates` を import せず、テストではモックに差し替えられます。

```go
// internal/usecase/send_email_confirmation.go
// UseCase 側で interface を定義（templates に依存しない）
type EmailConfirmationSender interface {
    Send(ctx context.Context, to, code, appURL, locale string) error
}

type SendEmailConfirmationUsecase struct {
    sender EmailConfirmationSender
}

func (uc *SendEmailConfirmationUsecase) Execute(ctx context.Context, input SendEmailConfirmationInput) error {
    return uc.sender.Send(ctx, input.Email, input.Code, input.AppURL, input.Locale)
}
```

`main.go` で `email.ConfirmationSender` を UseCase に注入します。

**依存の方向**:

```
Worker (Presentation)       → UseCase
UseCase (Application)       → EmailConfirmationSender (interface、UseCase 側で定義)
email.ConfirmationSender    → Sender, templates, i18n
```

UseCase は `templates` を直接 import しない。テンプレートレンダリングは email パッケージに閉じる。

### depguard ルール

email パッケージは以下の依存関係ルールに従います。

**許可される依存先**: `templates`、`i18n`、`model`、`config`、外部パッケージ（`templ`、`resend`）

**禁止される依存先**: `query`、`repository`、`handler`、`middleware`、`viewmodel`、`usecase`、`validator`、`worker`、`dispatcher`、`session`

## 認可チェック（Policy）

### 概要

認可チェック（`policy.TopicPolicy` など）は **UseCase** 内で行う。UseCase がオーケストレーターとして、データ取得後に認可チェックを実行する。

### 方針

- UseCase がデータ取得・認可チェック・バリデーション・永続化を統括する
- UseCase 内で Policy を呼び出して認可チェックを実行する
- Handler から Policy への直接依存は depguard で禁止する

### 理由

- エントリーポイントが増えた場合（Web API など）でも、認可チェックの漏れが発生しない
- 認可・バリデーション・ビジネスロジックが UseCase に集約され、一貫性が保たれる
- Handler は HTTP の入出力変換に専念でき、ドメイン固有の判断を持たない

### 実装パターン

```go
// UseCase 内で認可チェックを実行
func (uc *UpdateSuggestionUsecase) Execute(ctx context.Context, input UpdateSuggestionInput) (*UpdateSuggestionOutput, error) {
    // 1. データ取得
    space, err := uc.spaceRepo.FindByIdentifier(ctx, input.SpaceIdentifier)
    // ...
    spaceMember, err := uc.spaceMemberRepo.Find(ctx, space.ID, input.UserID)
    // ...

    // 2. 認可チェック
    if !policy.NewTopicPolicy(spaceMember, topicMember).CanUpdateSuggestion(suggestion) {
        return nil, &model.AppError{
            Code:    model.AppErrCodeForbidden,
            UserMsg: i18n.T(ctx, "error_forbidden"),
        }
    }

    // 3. バリデーション
    // 4. 永続化
    // ...
}
```

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

- **Handler / Worker**: 薄い Adapter（HTTP/ジョブの入出力変換のみ）
- **UseCase**: オーケストレーター（認可・バリデーション・ビジネスロジック・永続化を統括）
- **Validator**: バリデーション（`internal/validator/` に全て配置、UseCase から呼び出される）
- **Policy**: 認可ロジック（UseCase から呼び出される）
- **Dispatcher**: ジョブキューへの投入を抽象化（UseCase から呼び出される）
- **ViewModel**: リポジトリ層のデータをテンプレート表示用に変換

この構造により、エントリーポイントが増えても認可・バリデーションが漏れない、テストしやすく保守しやすいアーキテクチャを実現できます。

## 採用しなかった方針

### A. Handler から Repository への直接依存を許可する

Handler が Repository を直接呼び出すことを許可し、規約とコードレビューで書き込み呼び出しを防止する方針。

**不採用の理由**:

- 規約だけでは書き込みメソッドの呼び出しを防止できず、実際に違反が発生した
- Handler の依存先が UseCase と Repository の 2 つに分散し、ルールが複雑になる
- 依存グラフが一方向に統一されず、アーキテクチャの見通しが悪い

### B. Repository を ReadRepository と WriteRepository に分離する

Repository をインターフェースで分離し、Handler には ReadRepository のみ注入する方針。

**不採用の理由**:

- インターフェースの管理コストが増える
- Go のプラグマティックな哲学に反する（過度な抽象化）
- UseCase 経由に統一するほうがルールとしてシンプル

### C. Validator を handler パッケージ内に残す

Validator を `internal/handler/` 内に配置したまま、depguard による強制は諦めて構造的な規約とコードレビューで対応する方針。

**不採用の理由**:

- depguard で強制できないと、Handler から Repository への依存違反が再発する可能性がある
- 「Handler パッケージは repository を import しない」というルールを完全に強制できるメリットが大きい
- Validator を独立パッケージにすることで、将来的に Worker からもバリデーションを再利用できる

### D. ジョブの enqueue を Repository に含める

Repository がデータベースの抽象化層であるため、ジョブキューへの投入も Repository に含める方針。

**不採用の理由**:

- Repository は同期的なデータ永続化（CRUD）を担当し、Model と 1:1 で対応する。ジョブキューへの投入（非同期タスク委譲）は操作の性質が異なる
- Repository は `WithTx` パターンでトランザクションに参加するが、ジョブ投入はトランザクションとは別のライフサイクルを持つ
- `EnqueueEmailConfirmation` をどの Repository に置くかという判断コストが発生する
- 分離の基準はインフラの種類（PostgreSQL vs Redis vs River）ではなく、操作の性質（同期的データ永続化 vs 非同期タスク委譲）とする

### E. ValidationError と AppError を Application 層に配置する

`internal/usecase/errors.go` または新設の `internal/apperror/` にエラー型を定義する方針。

**不採用の理由**:

- Validator（Application 層）が `ValidationError` を生成するために `usecase` パッケージを import すると、UseCase → Validator という依存の方向に対して Validator → UseCase の逆方向依存が発生し、循環依存のリスクがある
- 新設パッケージ（`internal/apperror/`）を作ると、パッケージが増えて複雑になる
- Model（Domain/Infrastructure 層）は依存グラフの最下層にあり、すべての層から自然に参照できるため、エラー型の配置先として適切

### F. Worker（Presentation 層）でテンプレートをレンダリングし UseCase に渡す

Worker がメールテンプレートをレンダリングし、レンダリング済み HTML を UseCase に渡す方針。Handler がテンプレートを描画するのと同じパターンで、アーキテクチャの例外が不要という利点があった。

**不採用の理由**:

- 将来テンプレート選択にビジネスロジック（例: ユーザーのプランに応じた内容の分岐）が必要になった場合、Worker 側でその判断ができない
- その時点で UseCase 内でレンダリングに変更する手間が発生する
- 最初から UseCase にレンダリングを配置しておけば、判断コストが不要

### G. メールテンプレートを独立パッケージに分離する

メールテンプレートを `internal/email/templates/` のような独立パッケージに移動する方針。

**不採用の理由**:

- パッケージが増えて複雑になる
- メールテンプレートも templ で記述しており、HTTP レスポンス用テンプレートと同じツールチェーンを使用しているため、分離するメリットが薄い

### H. Validator を Domain/Infrastructure 層に配置する

Validator は入力のバリデーションだけでなく Repository に依存した状態バリデーションも行うため、Domain/Infrastructure 層に配置する方針。DDD やクリーンアーキテクチャでは、ドメインの不変条件やビジネスルールの検証はドメイン層に属するという考え方に基づく。

**不採用の理由**:

- このプロジェクトでは Domain 層と Infrastructure 層を統合しているため、Validator を Domain/Infrastructure 層に置くと Query（sqlc 生成コード）や Dispatcher（ジョブキュー）と同じ層に属することになり、バリデーションの性質とは異なる要素と混在する
- Validator は UseCase のオーケストレーションの一部（データ取得 → 認可 → バリデーション → 永続化）として設計されており、UseCase と同じ Application 層に配置することでこの関係が明確になる
- このプロジェクトの Validator は「フォーム入力の検証 + ユースケース固有の状態検証」という性格が強く、DDD でいう「ドメインの不変条件」（値オブジェクトやエンティティが常に満たすべき条件）とは役割が異なる

**補足**: Handler から Validator への直接依存の禁止は、Validator がどの層にいても depguard の設定で制御可能であり、層の選択理由にはならない。DDD の文脈でドメインの不変条件を表現したい場合は、Validator の配置を変えるのではなく、値オブジェクト（例: `EmailAddress` 型）を Model 層に導入するのが適切なアプローチである

### I. Templates から Model に直接依存して型安全性を確保する

Templates のパスヘルパー関数（`PagePath` など）の引数にドメイン ID 型（`model.PageNumber` など）を使用し、型安全性を高める方針。Templates は既に ViewModel を経由して間接的に Model に依存しており、依存の方向も Presentation 層 → Domain/Infrastructure 層で正しいため、直接依存を許可しても問題ないという考え方に基づく。

**不採用の理由**:

- depguard では「Model の型エイリアスのみ許可し、Model の構造体は禁止する」という粒度の制御ができない。Templates が Model に依存できるようになると、`model.Page` のようなモデル構造体をテンプレートに直接渡すミスを静的解析で検出できなくなる
- ViewModel による変換層のバイパスをコードレビューだけで防ぐのは漏れが発生しやすい

**代替として採用した方針**: ViewModel パッケージに Presentation 層用の型を定義する（例: `type PageNumber model.PageNumber`）。Templates は ViewModel の型のみに依存し、depguard による境界の強制を維持する。ViewModel のコンストラクタで `model.PageNumber` → `viewmodel.PageNumber` の変換を行う
