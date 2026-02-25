# Repository層の導入とアーキテクチャリファクタリング 設計書

<!--
このテンプレートの使い方:
1. このファイルを `.claude/designs/2_todo/` ディレクトリにコピー
   例: cp .claude/designs/template.md .claude/designs/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残しておくことを推奨
-->

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

現在のプロジェクトでは、sqlcで生成されたコードを`internal/repository/sqlc/`に配置し、Handlerから直接呼び出しています。しかし、sqlcで生成されるクエリは基本的に単一テーブルの操作であり、複数テーブルをまたがるデータ取得やトランザクション管理の責務を持つ「Repository」としては適切ではありません。

また、現状ではRepositoryがViewModelに依存するという、Layered Architectureの原則に反する設計になってしまう恐れがあります。

このリファクタリングでは、sqlc生成コードを`internal/query/`に移動し、**3層アーキテクチャ**（Presentation層、Application層、Domain/Infrastructure層）を採用します。Domain層とInfrastructure層は統合して考え、実用的でシンプルな設計を目指します。

**3層アーキテクチャを採用する理由**:

- **実用的**: データベース変更（PostgreSQL → MySQLなど）はほぼ起こらない
- **シンプル**: Domain層とInfrastructure層を分けることで生じる変換コストを削減
- **Goらしい**: Goのプラグマティックな哲学に合致
- **YAGNI原則**: 必要になってから層を分ければ良い

**目的**:

- 正しい依存関係を保つ（Presentation層 → Domain/Infrastructure層）
- 関心の分離を明確にする（Query = SQL実行、Repository = Query結果をModelに変換、Model = ドメインエンティティ、ViewModel = 表示用データ構造）
- ModelとRepositoryを1:1の関係で作成し、一貫性を保つ（同じ層なので依存できる）
- 特定クエリへの依存を排除し、再利用性を向上させる
- Handlerのコードを簡潔にし、可読性を向上させる
- テストの容易性を向上させる

**背景**:

- 現在のHandlerでは複数のクエリ呼び出しとデータ組み合わせロジックが散在している（例: `popular_work/index.go`）
- sqlcで生成されるコードは「クエリ実行層」であり、「Repository」という名称は責務を正しく表していない
- RepositoryがViewModelに依存すると、下位層が上位層に依存する逆転が発生する
- ModelとRepositoryは同じ層（Domain/Infrastructure層）として扱うことで、依存関係がシンプルになる

## 要件

<!--
ガイドライン:
- 機能要件: 「何ができるべきか」を記述
- 非機能要件: 「どのように動くべきか」を必要に応じて記述
-->

### 機能要件

<!--
「ユーザーは〇〇できる」「システムは〇〇する」という形式で記述
箇条書きで簡潔に
-->

- sqlcで生成されるコードは`internal/query/`パッケージに配置される（Domain/Infrastructure層）
- Model層は`internal/model/`パッケージに配置され、ページに依存しない汎用的なドメインエンティティを定義する（Domain/Infrastructure層）
- Repository層は`internal/repository/`パッケージに配置され、Query結果をModelに変換する（Domain/Infrastructure層）
- **ModelとRepositoryは同じ層**（Domain/Infrastructure層）に属するため、相互に依存できる
- **ModelとRepositoryは1:1の関係**で作成する（例: `model.Work` ↔ `repository.WorkRepository`）
- Repository層は複数のクエリを組み合わせて、ビジネスロジックに必要なModel構造を返す
- HandlerはRepositoryを呼び出してModelを取得し、ModelをViewModelに変換する（Presentation層）
- ViewModel層は引き続き表示用データ構造として維持され、Presentation層のヘルパー（`image.Helper`など）に依存できる
- **依存関係の方向**: Presentation層 → Application層 → Domain/Infrastructure層
- **既存のApplication層（UseCase）を維持**: `internal/usecase/`は既に存在し、ビジネスフロー・トランザクション管理を担当
- 既存の機能は変更せず、内部構造のみリファクタリングする

### 非機能要件

<!--
必要に応じて以下のような項目を追加してください：
- セキュリティ（認証、認可、暗号化、監査ログなど）
- パフォーマンス（応答時間、スループット、リソース使用量など）
- ユーザビリティ（UX）（使いやすさ、わかりやすさ、アクセシビリティなど）
- 可用性・信頼性（稼働率、障害時の挙動、エラーハンドリングなど）
- 保守性（テストのしやすさ、コードの読みやすさ、ドキュメントなど）

不要な場合はこのセクション全体を削除してください。
-->

- **保守性**: Repository層の導入により、データアクセスロジックの再利用性が向上する
- **テストのしやすさ**: Repository層をモック可能にすることで、Handlerのテストが容易になる
- **パフォーマンス**: 既存のパフォーマンスを維持する（N+1クエリ問題の回避など）
- **後方互換性**: リファクタリングによって既存の機能が壊れないこと

## 設計

<!--
ガイドライン:
- 技術的な実装の詳細を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - 技術スタック（使用するライブラリ、フレームワーク、ツールなど）
  - アーキテクチャ（システム全体の構成、コンポーネント間の関係など）
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - セキュリティ設計（認証・認可、トークン管理、Rate Limitingなど）
  - コード設計（パッケージ構成、主要な構造体、インターフェースなど）
  - テスト戦略（単体テスト、統合テスト、E2Eテストの方針）
  - マイグレーション管理（データベースマイグレーションの方針）
  - 実装方針（特記事項、既存システムとの関係、制約など）

不要な場合はこのセクション全体を削除してください。
-->

### アーキテクチャ

**変更前**:

```
sqlc → Handler → ViewModel → Template
```

**変更後（3層アーキテクチャ）**:

```
Query → Repository → Model → Handler → ViewModel → Template
  ↑         ↑          ↑        ↑          ↑
     Domain/Infrastructure層      Presentation層
     （統合して考える）
```

**3層アーキテクチャの構成**:

```
┌─────────────────────────────────────┐
│ Presentation層                       │
│ - Handler                           │
│ - ViewModel                         │
│ - Template                          │
└─────────────────────────────────────┘
         ↓ 依存（OK）
┌─────────────────────────────────────┐
│ Application層（既存）                 │
│ - UseCase（internal/usecase/）      │
│ - ビジネスフロー、トランザクション管理 │
└─────────────────────────────────────┘
         ↓ 依存（OK）
┌─────────────────────────────────────┐
│ Domain/Infrastructure層              │
│ - Query (sqlc)                      │
│ - Repository                        │
│ - Model                             │
│ （同じ層なので相互に依存できる）       │
└─────────────────────────────────────┘
```

**重要**: Domain/Infrastructure層はPresentation層に**依存してはいけない**

**Domain/Infrastructure層を統合する理由**:

- データベース変更（PostgreSQL → MySQLなど）は実際にはほぼ起こらない
- 層をまたぐ変換コストを削減し、シンプルさを保つ
- RepositoryとModelを同じ層として扱うことで、依存関係が自然になる
- 必要になったら分ければ良い（YAGNI原則）

**データの流れ**:

1. **Query** (Domain/Infrastructure層): SQLクエリを実行し、クエリ結果（`query.GetPopularWorksRow`など）を返す
2. **Repository** (Domain/Infrastructure層): Query結果をModelに変換し、複数のクエリを組み合わせる
3. **Model** (Domain/Infrastructure層): ページに依存しない汎用的なドメインエンティティ（`model.Work`など）
4. **Handler** (Presentation層): RepositoryからModelを取得し、ModelをViewModelに変換
5. **ViewModel** (Presentation層): 表示用のデータ構造（画像URL生成、言語切り替えなど）
6. **Template** (Presentation層): ViewModelを受け取ってHTMLを生成

**ModelとRepositoryの1:1関係**（同じ層なので依存できる）:

- `model.Work` ↔ `repository.WorkRepository`
- `model.User` ↔ `repository.UserRepository`
- `model.Episode` ↔ `repository.EpisodeRepository`

### ディレクトリ構造

**変更前**:

```
internal/
├── repository/
│   ├── queries/         # SQLクエリファイル
│   ├── sqlc/            # sqlc生成コード
│   │   ├── db.go
│   │   ├── models.go
│   │   ├── querier.go
│   │   └── works.sql.go
│   ├── sign_in_codes_test.go
│   └── sign_up_codes_test.go
├── viewmodel/
│   ├── work.go
│   └── page_meta.go
└── handler/
    └── popular_work/
        ├── handler.go
        └── index.go
```

**変更後**:

```
internal/
├── query/               # sqlc生成コード（旧 repository/sqlc）
│   ├── queries/         # SQLクエリファイル
│   ├── db.go
│   ├── models.go
│   ├── querier.go
│   └── works.sql.go
├── model/               # ドメインモデル（新規）
│   ├── work.go          # 作品のドメインモデル（Work, WorkWithDetails）
│   ├── cast.go          # キャストのドメインモデル（Cast）
│   └── staff.go         # スタッフのドメインモデル（Staff）
├── repository/          # Repository層（新規）
│   ├── work.go          # WorkRepository（Query→Model変換）
│   ├── work_test.go
│   ├── user.go          # UserRepository
│   └── user_test.go
├── viewmodel/           # 表示用データ構造（そのまま維持）
│   ├── work.go          # Model→Viewmodel変換
│   └── page_meta.go
└── handler/
    └── popular_work/
        ├── handler.go
        └── index.go
```

**注**: 上記はリファクタリング対象のディレクトリのみを表示しています。実際の`internal/`配下には、他にも以下のようなパッケージが存在します：

**Presentation層のヘルパー**:

- `image/` - 画像URL生成（imgproxy署名付きURL）
- `i18n/` - 国際化（翻訳取得、言語切り替え）
- `session/` - セッション管理（フラッシュメッセージ、ユーザー情報）
- `middleware/` - HTTPミドルウェア
- `templates/` - templテンプレート

**Application層**:

- `usecase/` - ユースケース（既存）

**その他**:

- `config/` - 設定管理
- `auth/` - 認証ロジック
- `turnstile/` - Cloudflare Turnstile連携

**物理的な構造と論理的な構造**:

- **物理的な構造**: `internal/`配下はフラット（機能別にパッケージを分ける）
- **論理的な構造**: ドキュメントでレイヤーごとにパッケージを分類し、依存関係を明示

### パッケージ間の依存関係

#### 基本方針

**レイヤー間の依存**: Presentation層 → Application層 → Domain/Infrastructure層

**レイヤー内の依存**: 各レイヤー内でも明確なルールに従ってパッケージ間の依存関係を管理

#### Presentation層内の依存関係

```
┌─────────────────────────────────────┐
│ Presentation層                       │
│                                     │
│  Handler ────→ ViewModel            │
│    │              │                 │
│    │              └→ image.Helper   │
│    │              └→ i18n           │
│    │              └→ session        │
│    │                                │
│    ├─→ UseCase（複雑なフロー）       │
│    └─→ Repository（データ取得）      │
│                                     │
│  Middleware ──→ session             │
└─────────────────────────────────────┘
```

**ルール**:

- HandlerはQuery/Modelに**直接依存しない**
- HandlerはRepositoryまたはUseCaseを経由してデータを取得
- ViewModelはPresentation層のヘルパー（image, i18n, session）に依存できる

#### Application層内の依存関係

```
┌─────────────────────────────────────┐
│ Application層                        │
│                                     │
│  UseCase ────→ Repository           │
│    │              │                 │
│    │              └→ Model（返却値）│
│    │                                │
│    └─→ Model（直接使用も可）         │
└─────────────────────────────────────┘
```

**ルール**:

- UseCaseはQuery/ViewModelに**直接依存しない**
- UseCaseはRepositoryを経由してデータアクセス
- UseCaseはModelを受け取り、Modelを返す

#### Domain/Infrastructure層内の依存関係

```
┌─────────────────────────────────────┐
│ Domain/Infrastructure層              │
│                                     │
│  Repository ──→ Query               │
│       │            │                │
│       └───→ Model（変換結果）        │
│                                     │
│  Model ←─────── Repository          │
│  （依存される側）                     │
└─────────────────────────────────────┘
```

**ルール**:

- **RepositoryのみがQueryに依存**（最重要）
- RepositoryはQuery結果をModelに変換
- ModelはPresentation層に**依存しない**

#### 全体の依存関係図

```
┌─────────────────────────────────────────────────────────┐
│ Presentation層                                           │
│                                                         │
│  Handler ──→ ViewModel ──→ (image, i18n, session)      │
│     │                                                   │
│     ├──→ UseCase ─────┐                                │
│     └──→ Repository ──┼─→ Model                        │
│                       │                                 │
└───────────────────────┼─────────────────────────────────┘
                        ↓
┌───────────────────────┼─────────────────────────────────┐
│ Application層         │                                 │
│                       │                                 │
│     ┌──← UseCase ←────┘                                │
│     │      │                                            │
│     │      └──→ Repository ──→ Model                   │
│     │                                                   │
└─────┼─────────────────────────────────────────────────┘
      ↓
┌─────┼─────────────────────────────────────────────────┐
│ Domain/Infrastructure層                                │
│     │                                                  │
│     └──→ Repository ──→ Query                          │
│              │                                         │
│              └──→ Model                                │
│                                                        │
└────────────────────────────────────────────────────────┘
```

#### 許可される依存

| 依存元     | 依存先                      | 理由                         |
| ---------- | --------------------------- | ---------------------------- |
| Handler    | Repository                  | データ取得の標準的な方法     |
| Handler    | UseCase                     | 複雑なビジネスフローの実行   |
| Handler    | ViewModel                   | Presentation層内の変換       |
| UseCase    | Repository                  | データアクセスの標準的な方法 |
| UseCase    | Model                       | ビジネスロジックで使用       |
| Repository | Query                       | データベースクエリの実行     |
| Repository | Model                       | Query結果の変換              |
| ViewModel  | Model                       | 表示用データへの変換         |
| ViewModel  | image.Helper, i18n, session | Presentation層のヘルパー     |

#### 禁止される依存

| 依存元     | 依存先                      | 理由                                 |
| ---------- | --------------------------- | ------------------------------------ |
| Handler    | Query                       | データアクセスロジックの散在を避ける |
| Handler    | Model                       | Repositoryを経由すべき               |
| UseCase    | Query                       | データアクセスロジックの散在を避ける |
| UseCase    | ViewModel                   | 上位層への依存（逆方向）             |
| Repository | ViewModel                   | 上位層への依存（逆方向）             |
| Model      | image.Helper, i18n, session | Presentation層への依存（逆方向）     |

#### 重要なルール

1. **Queryへの依存はRepositoryのみ**: Handler/UseCaseがQueryに直接依存することは禁止
2. **すべてのデータアクセスはRepositoryを経由**: データ取得はRepositoryまたはUseCaseを使う
3. **下位層は上位層に依存しない**: Domain/Infrastructure層はPresentation層に依存しない
4. **関心の分離**: 各パッケージは明確な責務を持ち、その責務に集中する

#### なぜRepositoryのみがQueryに依存すべきか

**メリット**:

- ✅ **保守性**: データアクセスロジックがRepositoryに集約される
- ✅ **拡張性**: キャッシュ層の追加、データソース変更がRepositoryのみで完結
- ✅ **一貫性**: 「データ取得 = Repositoryを使う」というルールが明確
- ✅ **テスト容易性**: Repositoryをモックすれば、Handler/UseCaseのテストが容易

**デメリットを回避**:

- ❌ データアクセスロジックの散在（Handler/UseCaseに直接Queryを書く）
- ❌ 変更の波及（データアクセス方法の変更がHandler/UseCaseに影響）
- ❌ ルールの曖昧さ（「このケースはQueryを直接使って良い？」という混乱）

### パッケージの責務

#### `internal/query/`（Domain/Infrastructure層）

- sqlcで自動生成されるコード
- 単一のSQLクエリを実行する責務のみ
- 手動編集禁止
- **例**: `query.GetPopularWorksRow`、`query.GetCastsByWorkIDsRow`

#### `internal/model/`（Domain/Infrastructure層）

- ページに依存しない汎用的なドメインエンティティ
- 作品、キャスト、スタッフなどのビジネスエンティティを表現
- **Presentation層に依存しない**（`image.Helper`などに依存しない）
- **例**: `model.Work`、`model.Cast`、`model.Staff`、`model.WorkWithDetails`
- 複雑なビジネスロジックは含まない（将来的に必要になったら追加）
- **ModelとRepositoryは1:1の関係**

#### `internal/repository/`（Domain/Infrastructure層）

- Query結果をModelに変換する
- 複数のクエリを組み合わせてModelを構築
- トランザクション内でのクエリ実行（トランザクションのライフサイクル管理はUseCaseが担当）
- **Modelと同じ層（Domain/Infrastructure層）なので、相互に依存できる**
- **Presentation層（ViewModel）には依存しない**
- **例**: `WorkRepository.GetPopularWorksWithDetails()` は `[]model.WorkWithDetails` を返す
- **ModelとRepositoryは1:1の関係**

#### `internal/viewmodel/`（Presentation層）

- Modelをテンプレート表示用のデータ構造に変換
- 国際化対応（言語切り替え）
- 画像URL生成などの表示ロジック
- **Presentation層のヘルパー（`image.Helper`など）に依存できる**
- **例**: `viewmodel.NewWorksFromModelDetails(models, imageHelper)`

#### `internal/handler/`（Presentation層）

- HTTPリクエスト処理
- RepositoryからModelを取得（単純な取得の場合）
- **UseCaseを呼び出す**（複雑なビジネスフローの場合）
- **ModelをViewModelに変換**（Presentation層内の変換）
- ViewModelをTemplateに渡す

#### `internal/usecase/`（Application層 - 既存）

**既に存在**しており、以下の責務を担当：

- **ビジネスフロー**: 複数のRepositoryを組み合わせた処理
- **トランザクションのライフサイクル管理**: トランザクション開始（`db.BeginTx()`）、コミット（`tx.Commit()`）、ロールバック（`tx.Rollback()`）
- **ビジネスルール**: セッション作成、パスワードリセット、サインアップ/サインインなど
- **例**:
  - `CreateSessionUsecase`: セッション作成のビジネスロジック
  - `CreatePasswordResetTokenUsecase`: パスワードリセットトークン生成
  - `UpdatePasswordResetUsecase`: パスワードリセット処理
  - `CompleteSignUpUsecase`: サインアップ完了処理

**今回のリファクタリングでの扱い**:

- 既存のUseCaseはそのまま維持
- 新しく追加するModel/Repository層はUseCaseからも呼び出せる
- Repositoryにトランザクション（`tx`）を渡して、トランザクション内でクエリを実行させる

### コード設計

#### Model層の実装例

```go
// internal/model/work.go
package model

import "time"

// Work は作品のドメインエンティティです（ページに依存しない汎用的な構造）
// Domain/Infrastructure層に属し、Presentation層に依存しない
type Work struct {
    ID                 int64
    Title              string
    TitleEn            string
    TitleKana          *string
    RecommendedImageURL string
    ImageData          string  // work_imagesテーブルのimage_data (JSON)
    WatchersCount      int32
    SeasonYear         *int32
    SeasonName         *int32  // シーズン番号（0=冬、1=春、2=夏、3=秋）
    CreatedAt          time.Time
}

// WorkWithDetails は作品の詳細情報を含むデータ構造です
type WorkWithDetails struct {
    Work   Work
    Casts  []Cast
    Staffs []Staff
}
```

```go
// internal/model/cast.go
package model

// Cast はキャスト情報を表します
type Cast struct {
    ID              int64
    WorkID          int64
    Name            string
    NameEn          string
    CharacterName   string
    CharacterNameEn string
    PersonName      string
    PersonNameEn    string
}
```

```go
// internal/model/staff.go
package model

// Staff はスタッフ情報を表します
type Staff struct {
    ID          int64
    WorkID      int64
    Name        string
    NameEn      string
    Role        string
    RoleOther   string
    RoleOtherEn string
}
```

**重要**: Model層はPresentation層に依存しない（`image.Helper`などに依存しない）

#### Repository層の実装例

```go
// internal/repository/work.go
package repository

import (
    "context"
    "github.com/annict/annict/internal/model"
    "github.com/annict/annict/internal/query"
)

// WorkRepository はWork関連のデータアクセスを担当します
type WorkRepository struct {
    queries *query.Queries
}

// NewWorkRepository はWorkRepositoryを作成します
func NewWorkRepository(queries *query.Queries) *WorkRepository {
    return &WorkRepository{queries: queries}
}

// GetPopularWorksWithDetails は人気作品をキャスト・スタッフ情報と共に取得します
func (r *WorkRepository) GetPopularWorksWithDetails(ctx context.Context) ([]model.WorkWithDetails, error) {
    // 1. クエリ実行
    worksRows, err := r.queries.GetPopularWorks(ctx)
    if err != nil {
        return nil, err
    }

    if len(worksRows) == 0 {
        return []model.WorkWithDetails{}, nil
    }

    // 2. query.GetPopularWorksRow → model.Work に変換
    works := make([]model.Work, len(worksRows))
    workIDs := make([]int64, len(worksRows))
    for i, row := range worksRows {
        works[i] = r.workFromPopularRow(row)
        workIDs[i] = row.ID
    }

    // 3. キャストとスタッフを取得
    castsRows, err := r.queries.GetCastsByWorkIDs(ctx, workIDs)
    if err != nil {
        return nil, err
    }

    staffsRows, err := r.queries.GetStaffsByWorkIDs(ctx, workIDs)
    if err != nil {
        return nil, err
    }

    // 4. query結果をmodelに変換
    casts := r.castsFromRows(castsRows)
    staffs := r.staffsFromRows(staffsRows)

    // 5. 組み合わせる
    return r.combineWorkData(works, casts, staffs), nil
}

// workFromPopularRow は query.GetPopularWorksRow を model.Work に変換します
func (r *WorkRepository) workFromPopularRow(row query.GetPopularWorksRow) model.Work {
    work := model.Work{
        ID:                 row.ID,
        Title:              row.Title,
        TitleEn:            row.TitleEn,
        RecommendedImageURL: row.RecommendedImageUrl,
        ImageData:          row.ImageData.String,
        WatchersCount:      row.WatchersCount,
        CreatedAt:          row.CreatedAt,
    }

    if row.SeasonYear.Valid {
        work.SeasonYear = &row.SeasonYear.Int32
    }
    if row.SeasonName.Valid {
        work.SeasonName = &row.SeasonName.Int32
    }

    return work
}

// castsFromRows は query結果を model.Cast に変換します
func (r *WorkRepository) castsFromRows(rows []query.GetCastsByWorkIDsRow) []model.Cast {
    casts := make([]model.Cast, len(rows))
    for i, row := range rows {
        casts[i] = model.Cast{
            ID:              row.ID,
            WorkID:          row.WorkID,
            Name:            row.Name,
            NameEn:          row.NameEn,
            CharacterName:   row.CharacterName.String,
            CharacterNameEn: row.CharacterNameEn.String,
            PersonName:      row.PersonName.String,
            PersonNameEn:    row.PersonNameEn.String,
        }
    }
    return casts
}

// staffsFromRows は query結果を model.Staff に変換します
func (r *WorkRepository) staffsFromRows(rows []query.GetStaffsByWorkIDsRow) []model.Staff {
    staffs := make([]model.Staff, len(rows))
    for i, row := range rows {
        staffs[i] = model.Staff{
            ID:          row.ID,
            WorkID:      row.WorkID,
            Name:        row.Name,
            NameEn:      row.NameEn,
            Role:        row.Role,
            RoleOther:   row.RoleOther.String,
            RoleOtherEn: row.RoleOtherEn,
        }
    }
    return staffs
}

// combineWorkData は作品データとキャスト・スタッフデータを組み合わせます
func (r *WorkRepository) combineWorkData(
    works []model.Work,
    casts []model.Cast,
    staffs []model.Staff,
) []model.WorkWithDetails {
    // キャストとスタッフをwork_idでマッピング
    castsMap := make(map[int64][]model.Cast)
    for _, cast := range casts {
        castsMap[cast.WorkID] = append(castsMap[cast.WorkID], cast)
    }

    staffsMap := make(map[int64][]model.Staff)
    for _, staff := range staffs {
        staffsMap[staff.WorkID] = append(staffsMap[staff.WorkID], staff)
    }

    // WorkWithDetailsのスライスを作成
    result := make([]model.WorkWithDetails, len(works))
    for i, work := range works {
        result[i] = model.WorkWithDetails{
            Work:   work,
            Casts:  castsMap[work.ID],
            Staffs: staffsMap[work.ID],
        }
    }

    return result
}
```

#### Viewmodel層の実装例

```go
// internal/viewmodel/work.go
package viewmodel

import (
    "github.com/annict/annict/internal/image"
    "github.com/annict/annict/internal/model"
)

// Work はテンプレート表示用の作品データです
type Work struct {
    ID            int64
    Title         string
    TitleEn       string
    ImageURL      string   // デフォルトの画像URL (280px, jpg)
    ImageDataJSON string   // work_imagesテーブルのimage_data (JSON)
    WatchersCount int32
    SeasonYear    *int32
    SeasonName    *string  // 表示用のシーズン名（日本語）
    SeasonNumber  *int32   // シーズン番号（翻訳キー用）
    Casts         []Cast
    Staffs        []Staff
    imageHelper   *image.Helper
}

// Cast はキャスト情報を表します
type Cast struct {
    ID              int64
    Name            string
    NameEn          string
    CharacterName   string
    CharacterNameEn string
    PersonName      string
    PersonNameEn    string
}

// Staff はスタッフ情報を表します
type Staff struct {
    ID          int64
    Name        string
    NameEn      string
    Role        string
    RoleOther   string
    RoleOtherEn string
}

// NewWorksFromModelDetails は model.WorkWithDetails から viewmodel.Work に変換します
func NewWorksFromModelDetails(details []model.WorkWithDetails, helper *image.Helper) []Work {
    works := make([]Work, len(details))
    for i, detail := range details {
        works[i] = NewWorkFromModelDetail(detail, helper)
    }
    return works
}

// NewWorkFromModelDetail は model.WorkWithDetails から viewmodel.Work に変換します
func NewWorkFromModelDetail(detail model.WorkWithDetails, helper *image.Helper) Work {
    // imgproxy用の画像URL生成（280pxサイズ、jpg形式）
    imageURL := ""
    if helper != nil {
        imageURL = helper.GetWorkImageURL(detail.Work.ImageData, 280, "jpg")
    }

    work := Work{
        ID:            detail.Work.ID,
        Title:         detail.Work.Title,
        TitleEn:       detail.Work.TitleEn,
        ImageURL:      imageURL,
        ImageDataJSON: detail.Work.ImageData,
        WatchersCount: detail.Work.WatchersCount,
        imageHelper:   helper,
    }

    // タイトルのフォールバック処理
    if work.Title == "" && detail.Work.TitleEn != "" {
        work.Title = detail.Work.TitleEn
    }

    // シーズン情報の変換
    if detail.Work.SeasonYear != nil {
        work.SeasonYear = detail.Work.SeasonYear
    }

    if detail.Work.SeasonName != nil {
        work.SeasonNumber = detail.Work.SeasonName
        // 日本語のシーズン名に変換
        seasonNames := []string{"冬", "春", "夏", "秋"}
        if *detail.Work.SeasonName >= 0 && *detail.Work.SeasonName < int32(len(seasonNames)) {
            seasonStr := seasonNames[*detail.Work.SeasonName]
            work.SeasonName = &seasonStr
        }
    }

    // キャストとスタッフの変換
    work.Casts = make([]Cast, len(detail.Casts))
    for i, cast := range detail.Casts {
        work.Casts[i] = Cast{
            ID:              cast.ID,
            Name:            cast.Name,
            NameEn:          cast.NameEn,
            CharacterName:   cast.CharacterName,
            CharacterNameEn: cast.CharacterNameEn,
            PersonName:      cast.PersonName,
            PersonNameEn:    cast.PersonNameEn,
        }
    }

    work.Staffs = make([]Staff, len(detail.Staffs))
    for i, staff := range detail.Staffs {
        work.Staffs[i] = Staff{
            ID:          staff.ID,
            Name:        staff.Name,
            NameEn:      staff.NameEn,
            Role:        staff.Role,
            RoleOther:   staff.RoleOther,
            RoleOtherEn: staff.RoleOtherEn,
        }
    }

    return work
}
```

#### Handlerの簡素化例

**変更前** (`internal/handler/popular_work/index.go`):

```go
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // リポジトリから人気作品を取得
    worksData, err := h.queries.GetPopularWorks(ctx)
    if err != nil {
        log.Printf("人気作品の取得エラー: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // ビューモデルに変換
    worksView := viewmodel.NewWorksFromPopularRows(worksData, h.imageHelper)

    // 作品IDのリストを作成
    workIDs := make([]int64, len(worksData))
    for i, work := range worksData {
        workIDs[i] = work.ID
    }

    // キャストデータを取得
    if len(workIDs) > 0 {
        casts, err := h.queries.GetCastsByWorkIDs(ctx, workIDs)
        if err != nil {
            log.Printf("キャスト情報の取得エラー: %v", err)
        } else {
            // キャストを作品IDでグループ化
            castsMap := viewmodel.GroupCastsByWorkID(casts)
            // 各作品にキャストを割り当て
            for i := range worksView {
                if c, exists := castsMap[worksView[i].ID]; exists {
                    worksView[i].Casts = c
                }
            }
        }

        // スタッフデータを取得
        staffs, err := h.queries.GetStaffsByWorkIDs(ctx, workIDs)
        if err != nil {
            log.Printf("スタッフ情報の取得エラー: %v", err)
        } else {
            // スタッフを作品IDでグループ化
            staffsMap := viewmodel.GroupStaffsByWorkID(staffs)
            // 各作品にスタッフを割り当て
            for i := range worksView {
                if s, exists := staffsMap[worksView[i].ID]; exists {
                    worksView[i].Staffs = s
                }
            }
        }
    }

    // ... テンプレートレンダリング
}
```

**変更後**:

```go
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. RepositoryからModelを取得（Infrastructure層 → Domain層）
    modelWorks, err := h.workRepo.GetPopularWorksWithDetails(ctx)
    if err != nil {
        log.Printf("人気作品の取得エラー: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // 2. ModelをViewModelに変換（Presentation層内の変換）
    viewWorks := viewmodel.NewWorksFromModelDetails(modelWorks, h.imageHelper)

    // 3. テンプレートにViewModelを渡す
    // ...
}
```

**ポイント**:

- HandlerがModel→ViewModel変換を行う（Presentation層内の変換）
- RepositoryはViewModelに依存しない（Domain/Infrastructure層がPresentation層に依存しない）
- RepositoryとModelは同じ層（Domain/Infrastructure層）なので、相互に依存できる

### テスト戦略

- **Query層**: sqlcが生成するため、基本的にテスト不要（既存のテストは維持）
- **Model層**: 単純なデータ構造のため、基本的にテスト不要（ビジネスロジックが追加されたらテストを追加）
- **Repository層**: 実データベースを使用した統合テストを実施
  - トランザクション分離により、各テストは独立して実行可能
  - `testutil.SetupTestDB(t)` を使用してテスト用DBをセットアップ
  - Query結果からModelへの変換ロジックをテスト
- **Viewmodel層**: ModelからViewmodelへの変換ロジックをユニットテスト
  - 画像URL生成、言語切り替えなどの表示ロジックをテスト
- **Handler層**: Repositoryをモック可能にすることで、単体テストが容易に

### 実装方針

1. **段階的な移行**: 既存の機能を壊さないよう、1つずつ移行する
2. **並行開発可能**: 既存のコードは動作し続けるため、新規機能開発と並行可能
3. **3層アーキテクチャ**: Presentation層、Application層（既存）、Domain/Infrastructure層（統合）
4. **正しい依存関係**: Presentation層 → Application層 → Domain/Infrastructure層
5. **フラットなディレクトリ構造**: `internal/`配下は機能別にパッケージを分ける（Goの標準的な考え方）
6. **論理的なレイヤー構造**: ドキュメントでレイヤーごとにパッケージを分類し、依存関係を明示
7. **ModelとRepositoryは同じ層**: 両者は同じDomain/Infrastructure層に属するため、相互に依存できる
8. **ModelとRepositoryは1:1の関係**: 各ドメインエンティティに対して対応するRepositoryを作成（例: `model.Work` ↔ `repository.WorkRepository`）
9. **Model層はシンプルに**: 複雑なビジネスロジックは含めず、必要になったら追加
10. **HandlerがModel→ViewModel変換を担当**: Presentation層内で変換を行う
11. **既存のUseCaseを維持**: Application層（`internal/usecase/`）は既に存在し、そのまま維持する
12. **UseCaseからもModel/Repositoryを利用可能**: 新しく追加するModel/Repository層はUseCaseからも呼び出せる
13. **後方互換性**: 既存のテストは継続して動作すること

## タスクリスト

<!--
ガイドライン:
- フェーズごとに段階的な実装計画を記述
- チェックボックスで進捗を管理
- **重要**: 1タスク = 1 Pull Request の粒度で作成してください
- **重要**: 各タスクには想定ファイル数と想定行数を明記してください（PRサイズの見積もりのため）
- 想定ファイル数は「実装」と「テスト」に分けて記載してください
- 想定行数も「実装」と「テスト」に分けて記載してください
- 依存関係を明確に
- Pull Requestのガイドラインは CLAUDE.md を参照（変更ファイル数20以下、変更行数300行以下）

タスク番号の付け方:
- 各タスクには階層的な番号を付与します（例: 1-1, 1-2, 2-1, 2-2）
- フォーマット: **フェーズ番号-タスク番号**: タスク名
- タスクの前に別のタスクを追加する場合は、サブ番号を使用します
  - 例: タスク 2-1 の前にタスクを追加する場合 → 2-0
  - 例: タスク 2-0 の前にタスクを追加する場合 → 2-0-1
- この番号はブランチ名の一部として使用されます（例: feature-1-1, feature-2-0）
-->

### フェーズ 1: Query層のリネームとsqlc設定変更

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: sqlcの出力先を`internal/query/`に変更
  - `sqlc.yaml`の`gen.go.out`を`internal/query`に変更
  - `sqlc generate`を実行して`internal/query/`配下にコード生成
  - 既存の`internal/repository/sqlc/`は一旦残しておく（後で削除）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 10 行（実装 10 行 + テスト 0 行）

- [x] **1-2**: import文を`internal/query`に変更
  - 全ファイルのimport文を`internal/repository/sqlc`から`internal/query`に変更
  - `go mod tidy`を実行
  - 既存のテストがすべてパスすることを確認
  - **想定ファイル数**: 約 30 ファイル（実装 30 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）※import文のみの変更

- [x] **1-3**: 旧`internal/repository/sqlc/`ディレクトリの削除
  - `internal/repository/sqlc/`ディレクトリを削除
  - `internal/repository/queries/`を`internal/query/queries/`に移動
  - `internal/repository/`配下の既存テストファイルを適切な場所に移動
  - **想定ファイル数**: 約 5 ファイル（実装 3 + テスト 2）
  - **想定行数**: 約 50 行（実装 20 行 + テスト 30 行）

### フェーズ 2: Model層とRepository層の導入（Work関連）

- [x] **2-1**: Model層の実装
  - `internal/model/work.go`を作成（`Work`、`WorkWithDetails`構造体）
  - `internal/model/cast.go`を作成（`Cast`構造体）
  - `internal/model/staff.go`を作成（`Staff`構造体）
  - ページに依存しない汎用的なデータ構造として設計
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 60 行（実装 60 行 + テスト 0 行）※データ構造のみ

- [x] **2-2**: WorkRepositoryの実装
  - `internal/repository/work.go`を作成
  - `WorkRepository`構造体と`NewWorkRepository`関数を実装
  - `GetPopularWorksWithDetails`メソッドを実装（Query→Model変換）
  - `workFromPopularRow`、`castsFromRows`、`staffsFromRows`、`combineWorkData`ヘルパー関数を実装
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 300 行（実装 150 行 + テスト 150 行）

- [x] **2-3**: ViewModelのリファクタリング
  - `viewmodel.NewWorksFromModelDetails`関数を追加（Model→ViewModel変換）
  - `viewmodel.NewWorkFromModelDetail`関数を追加
  - ViewModelは`image.Helper`に依存できる（Presentation層なので問題ない）
  - 既存の`NewWorksFromPopularRows`は後方互換性のため残す
  - `GroupCastsByWorkID`と`GroupStaffsByWorkID`は残す（popular_workハンドラーで使用中、タスク2-4で削除予定）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

- [x] **2-4**: popular_work HandlerでWorkRepositoryを使用
  - `Handler`構造体に`workRepo`フィールドを追加
  - `Index`メソッドを修正：
    1. `workRepo.GetPopularWorksWithDetails()`でModelを取得
    2. HandlerでModel→ViewModel変換を行う（`viewmodel.NewWorksFromModelDetails()`）
    3. ViewModelをテンプレートに渡す
  - 既存のテストを修正
  - **依存関係の確認**: RepositoryがViewModelに依存していないことを確認（Domain/Infrastructure層がPresentation層に依存しない）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 100 行（実装 50 行 + テスト 50 行）

### フェーズ 3: ドキュメント更新と旧コードのクリーンアップ

- [x] **3-1**: ドキュメントの更新
  - `go/CLAUDE.md`の「プロジェクト構造」セクションを更新
    - **レイヤーごとにパッケージを分類**（Presentation層、Application層、Domain/Infrastructure層）
    - **Presentation層のヘルパー一覧を追加**（`image`, `i18n`, `session`など）
    - **レイヤー間の依存関係を明示**（Presentation層 → Application層 → Domain/Infrastructure層）
    - **レイヤー内のパッケージ間依存関係を明示**（図とテーブルで説明）
    - **重要なルール**：「Queryへの依存はRepositoryのみ」「Handler/UseCaseはQueryに直接依存しない」
    - **物理的な構造（フラット）と論理的な構造（レイヤー）の関係を説明**
  - `go/docs/architecture-guide.md`を作成または更新
  - `.claude/designs/1_doing/go.md`のコード例を更新
  - **3層アーキテクチャの説明を追加**（Presentation層、Application層、Domain/Infrastructure層）
  - **Domain/Infrastructure層を統合する理由を明記**
  - **ModelとRepositoryの1:1関係を明記**
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 400 行（実装 400 行 + テスト 0 行）※依存関係図とルールを追加

- [x] **3-2**: 旧コードのクリーンアップ
  - `viewmodel`パッケージから不要な関数を削除
  - 未使用のimport文を削除
  - `go mod tidy`を実行
  - すべてのテストがパスすることを確認
  - **想定ファイル数**: 約 5 ファイル（実装 3 + テスト 2）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）※主に削除

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **Domain層とInfrastructure層の分離**: 3層アーキテクチャを採用し、両者を統合して扱う（YAGNI原則）
- **Model層への複雑なビジネスロジック追加**: 現時点では単純なデータ構造のみ定義し、ビジネスロジックは必要になってから追加
- **Repository層のインターフェース化**: 現時点ではモックが不要なため、具体的な構造体のみ実装
- **他の機能のRepository化**: まずはWork関連のみ実装し、効果を確認してから他の機能に展開
- **キャッシュ戦略の追加**: Repository層の導入後、必要に応じて検討

**将来的な拡張**:

- データベース変更が本当に必要になったら、その時にDomain層とInfrastructure層を分離

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [sqlc公式ドキュメント](https://docs.sqlc.dev/)
- [Go言語のクリーンアーキテクチャ](https://github.com/bxcodec/go-clean-arch)
- [Layered Architectureの依存関係](https://herbertograca.com/2017/08/03/layered-architecture/)
- [DDD (Domain-Driven Design) のレイヤードアーキテクチャ](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [YAGNI原則 (You Aren't Gonna Need It)](https://martinfowler.com/bliki/Yagni.html)
- [Goのプラグマティックな哲学](https://go-proverbs.github.io/)

---

## テンプレート使用例

実際の使用例は以下を参照してください：

- [パスワードリセット機能](../3_done/202510/password-reset.md)
