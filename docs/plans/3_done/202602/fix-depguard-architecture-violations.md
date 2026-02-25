# depguardアーキテクチャ違反の修正 作業計画書

<!--
このテンプレートの使い方:
1. このファイルを `docs/plans/2_todo/` ディレクトリにコピー
   例: cp docs/plans/template.md docs/plans/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残してください

**作業計画書の性質**:
- 作業計画書は「何をどう変えるか」という変更内容を記述するドキュメントです
- 新しい機能の場合は、概要・要件・設計もこのドキュメントに記述します
- 現在のシステムの状態は `docs/specs/` の仕様書に記述されています
- タスク完了後は、仕様書を新しい状態に更新してください（設計判断や採用しなかった方針も含める）

**仕様書との関係**:
- 新しい機能の場合: タスク完了後に `docs/specs/` に仕様書を作成する
- 既存機能の変更の場合: 「仕様書」セクションに対応する仕様書へのリンクを記載し、タスク完了後に仕様書を更新する

**公開時の注意事項**:
- 開発用ドメイン名を記載する場合は `example.dev` を使用してください（実際のドメイン名は記載しない）
- 環境変数の値はサンプル値のみ記載し、実際の値は含めないでください
-->

## 仕様書

<!--
- 既存機能を変更する場合: 変更対象の仕様書へのリンクを記載してください
- 新しい機能の場合: タスク完了後に作成予定の仕様書のパスを記載してください
-->

- [アーキテクチャガイド](../../go/docs/architecture-guide.md)

## 概要

<!--
ガイドライン:
- この機能が「何であるか」「なぜ必要か」を簡潔に説明
- 2-3段落程度で簡潔に
- 既存機能の変更の場合は、変更の背景と目的を記述
-->

Go CI の `make lint`（golangci-lint の depguard ルール）が4件のアーキテクチャ違反で失敗している。これらはプロジェクトの3層アーキテクチャルールに違反するimportであり、修正が必要。

違反の内訳:

1. **Application層 → Query層の直接依存（3件）**: `internal/usecase/` 配下の3ファイルが `internal/query` を直接importしている。アーキテクチャルールでは、Application層はRepositoryを経由してQuery層にアクセスする必要がある。
2. **ViewModel層 → Repository層の依存（1件）**: `internal/viewmodel/user.go` が `internal/repository` をimportしている。ViewModelはModelからの変換のみを行い、Repositoryに依存してはいけない。

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

- `make lint` が depguard 違反なしで成功すること
- 既存のテスト（`make test`）がすべてパスすること
- 既存の機能に変更がないこと（リファクタリングのみ）

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

- **保守性**: リファクタリング後のコードが既存のアーキテクチャパターンに準拠していること

## 実装ガイドラインの参照

<!--
**重要**: 作業計画書を作成する前に、対象プラットフォームのガイドラインを必ず確認してください。
特に以下の点に注意してください：
- ディレクトリ構造・ファイル名の命名規則
- コーディング規約
- アーキテクチャパターン

ガイドラインに沿わない設計は、実装時にそのまま実装されてしまうため、
作業計画書作成の段階でガイドラインに準拠していることを確認してください。
-->

### Go版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - 全体的なコーディング規約
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド

## 設計

<!--
ガイドライン:
- 技術的な実装の設計を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - UI設計（画面構成、ユーザーフローなど）
  - セキュリティ設計（認証・認可、トークン管理など）
  - コード設計（パッケージ構成、主要な構造体など）

**重要: 設計は実装中に更新する**:
- 作業計画書内の設計は初期の方針であり、完璧ではない
- 実装中により良いアプローチが見つかった場合は、設計を積極的に更新する
- 設計に固執して実装の質を下げるよりも、実装で得た知見を設計に反映する方が重要
- 変更した場合は「採用しなかった方針」セクションに変更前の方針と変更理由を記録する
-->

### 違反1: viewmodel/user.go の Repository依存

**現状**: `viewmodel/user.go` が `repository.User` 型を参照している。

```go
// 現在のコード
import "github.com/annict/annict/go/internal/repository"

func NewUserForSidebar(row *repository.User, helper *image.Helper) *User
```

**修正方針**: `repository.User` は `model.User` の型エイリアス（`type User = model.User`）なので、`model.User` を直接参照するように変更する。

```go
// 修正後
import "github.com/annict/annict/go/internal/model"

func NewUserForSidebar(row *model.User, helper *image.Helper) *User
```

`repository.User` = `model.User` なので、呼び出し側の変更は不要（型互換性あり）。

### 違反2: usecase/create_session.go の Query依存

**現状**: `CreateSessionUsecase` が `query.Queries` を直接保持し、`query.CreateSessionParams` を使ってセッションを作成している。

**修正方針**: 既存の `SessionRepository` には `CreateSession(ctx, sessionID, data)` メソッドがあるため、これを活用する。ただし、`CreateSessionUsecase` はprivate ID生成やセッションデータ構築のビジネスロジックを含むため、Repositoryへのデータアクセス委譲のみを行う。

```go
// 修正後
type CreateSessionUsecase struct {
    sessionRepo *repository.SessionRepository
}

func (uc *CreateSessionUsecase) Execute(ctx context.Context, tx *sql.Tx, ...) (*SessionResult, error) {
    // ビジネスロジック（public ID生成、セッションデータ構築）はそのまま
    // DB操作のみSessionRepositoryに委譲

    var sessionRepo *repository.SessionRepository
    if tx != nil {
        sessionRepo = uc.sessionRepo.WithTx(tx)
    } else {
        sessionRepo = uc.sessionRepo
    }

    // SessionRepositoryは内部でprivateID生成済みのCreateSessionを持つが、
    // usecaseが独自にprivateIDを生成しているため、新しいメソッドが必要
}
```

**注意点**: 現在の `SessionRepository.CreateSession()` はpublic IDからprivate IDへの変換を内部で行う。一方、`CreateSessionUsecase` も独自にprivate IDを生成している。この重複を解消するため、UseCase側のprivate ID生成ロジックを削除し、`SessionRepository.CreateSession()` に統一する。

具体的には:

- `CreateSessionUsecase` で `generatePrivateID()` を呼んでいた箇所を削除
- `SessionRepository.CreateSession(ctx, publicID, data)` を使用（private ID変換はRepository内部で実施）
- `SessionRepository` に `WithTx` メソッドを追加

### 違反3: usecase/create_password_reset_token.go の Query依存

**現状**: `CreatePasswordResetTokenUsecase` が `query.Queries` を直接保持し、`query.CreatePasswordResetTokenParams` でトークンを作成し、`queries.GetUserByID` でユーザー情報を取得している。

**修正方針**: 既存の `PasswordResetTokenRepository` にトークン作成・削除メソッドを追加し、ユーザー情報取得には `UserRepository` を使用する。

```go
// PasswordResetTokenRepository に追加するメソッド
func (r *PasswordResetTokenRepository) Create(ctx context.Context, userID int64, tokenDigest string, expiresAt time.Time) error
func (r *PasswordResetTokenRepository) DeleteUnusedByUserID(ctx context.Context, userID int64) error
func (r *PasswordResetTokenRepository) WithTx(tx *sql.Tx) *PasswordResetTokenRepository

// 修正後のUsecase
type CreatePasswordResetTokenUsecase struct {
    db                    *sql.DB
    userRepo              *repository.UserRepository
    passwordResetTokenRepo *repository.PasswordResetTokenRepository
    cfg                   *config.Config
    riverClient           *worker.Client
}
```

### 違反4: usecase/complete_sign_up.go の Query依存

**現状**: `CompleteSignUpUsecase` が `query.Queries` を直接保持し、以下のQuery操作を直接呼び出している:

- `queries.GetUserByUsername` - ユーザー名一意性チェック
- `queriesWithTx.CreateUser` - ユーザー作成
- `queriesWithTx.CreateProfile` - プロフィール作成
- `queriesWithTx.CreateSetting` - 設定作成
- `queriesWithTx.CreateEmailNotification` - メール通知設定作成
- `NewCreateSessionUsecase(uc.queries)` - セッション作成（他のUsecaseに委譲）

また、`CompleteSignUpResult.User` が `query.CreateUserRow` 型を使用している。

**修正方針**:

- `UserRepository` にユーザー作成と関連レコード作成メソッドを追加
- 新しいRepositoryとして `ProfileRepository`, `SettingRepository`, `EmailNotificationRepository` を作成するか、`UserRepository` に統合するかの判断が必要

`complete_sign_up.go` はユーザー登録のフローで以下を作成する: User → Profile → Setting → EmailNotification → Session。これらは密結合のため、`UserRepository` に `Create` メソッドを追加し、関連レコード作成は個別のRepositoryに委譲する方針とする。

```go
// UserRepository に追加
func (r *UserRepository) Create(ctx context.Context, params UserCreateParams) (*model.User, error)
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error)

// 新規Repository
// ProfileRepository
func (r *ProfileRepository) Create(ctx context.Context, userID int64, name string) error

// SettingRepository
func (r *SettingRepository) Create(ctx context.Context, userID int64) error

// EmailNotificationRepository
func (r *EmailNotificationRepository) Create(ctx context.Context, userID int64, unsubscriptionKey string) error

// CompleteSignUpResult の修正
type CompleteSignUpResult struct {
    User            *model.User  // query.CreateUserRow → model.User に変更
    SessionPublicID string
}
```

ハンドラー側（`sign_up_username/create.go`）で `ucResult.User.ID`, `ucResult.User.Username`, `ucResult.User.Email` を参照しているが、これらのフィールドは `model.User` にも存在するため、互換性がある。

### 変更による影響範囲

| 変更対象                                 | 影響を受けるファイル                                                                                                                                                   |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `viewmodel/user.go`                      | なし（型エイリアスのため互換性あり）                                                                                                                                   |
| `usecase/create_session.go`              | テスト: `create_session_test.go`、呼び出し元: `sign_in_code/handler.go`, `sign_in_password/handler.go`, `complete_sign_up.go`, `update_password_reset.go` + 関連テスト |
| `usecase/create_password_reset_token.go` | テスト: `create_password_reset_token_test.go`、呼び出し元: `password_reset/handler.go` + 関連テスト                                                                    |
| `usecase/complete_sign_up.go`            | テスト: `sign_up_username/create_test.go`、呼び出し元: `sign_up_username/handler.go` + 関連テスト                                                                      |

## 採用しなかった方針

<!--
ガイドライン:
- 検討したが採用しなかった設計や機能を、理由とともに記述
- 将来の開発者が同じ検討を繰り返さないための判断記録
- タスク完了後、この内容は `specs/` の仕様書にも転記する
- 該当がない場合は「なし」と記載
-->

### depguardルールの例外追加

UseCaseがQueryに直接アクセスすることをdepguardの例外として許可する方針も検討したが、アーキテクチャガイドに「Queryへの依存はRepositoryのみ」と明記されているため、コードをルールに合わせる方針とした。

### 全てのQuery型をmodel型でラップ

すべてのQuery戻り値型を `model` パッケージの型に変換する方針も検討したが、既存のRepositoryでは `query.GetUserByEmailForSignInRow` 等のQuery型をそのまま返しているケースが多数ある。今回はdepguard違反の修正に絞り、全体的なQuery型→Model型変換は将来のリファクタリングとする。

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
- **フェーズ番号は半角英数字とハイフンのみで表記**してください（ブランチ名に使用するため）
  - 例: フェーズ 1, フェーズ 2, フェーズ 5a（フェーズ 5 と 6 の間に追加する場合）
  - NG: フェーズ 5.5（ドットは使用不可）
- タスクの前に別のタスクを追加する場合は、サブ番号を使用します
  - 例: タスク 2-1 の前にタスクを追加する場合 → 2-0
  - 例: タスク 2-0 の前にタスクを追加する場合 → 2-0-1
- この番号はブランチ名の一部として使用されます（例: feature-1-1, feature-2-0）

プラットフォームプレフィックス:
- Go版またはRails版の修正を行うタスクには、タスク名の先頭にプラットフォームを示すプレフィックスを付けてください
- フォーマット: **フェーズ番号-タスク番号**: [Go] タスク名 または **フェーズ番号-タスク番号**: [Rails] タスク名
- Go版とRails版の両方を修正する場合は、別々のタスクに分けてください
- 例:
  - `- [ ] **1-1**: [Go] マイグレーション作成`
  - `- [ ] **1-2**: [Rails] モデルへのコールバック追加`
-->

### フェーズ 1: ViewModel層の修正

<!--
viewmodel/user.go の repository → model 依存変更。型エイリアスのため影響が小さく、最初に実施。
-->

- [x] **1-1**: [Go] viewmodel/user.go の Repository依存を Model依存に変更
  - `viewmodel/user.go` の import を `repository` → `model` に変更
  - `NewUserForSidebar(row *repository.User, ...)` → `NewUserForSidebar(row *model.User, ...)` に変更
  - `repository.User` は `model.User` の型エイリアスなので、呼び出し側の変更は不要
  - リント・テスト通過を確認
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 5 行（実装 5 行 + テスト 0 行）

### フェーズ 2: Repository層の拡張

<!--
UseCase修正の前提として、必要なRepositoryメソッドを追加する。
-->

- [x] **2-1**: [Go] SessionRepository に WithTx メソッドを追加
  - `repository/session.go` に `WithTx(tx *sql.Tx) *SessionRepository` メソッドを追加
  - テストを追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 10 行 + テスト 20 行）

- [x] **2-2**: [Go] PasswordResetTokenRepository にトークン作成・削除メソッドと WithTx を追加
  - `repository/password_reset_token.go` に以下を追加:
    - `WithTx(tx *sql.Tx) *PasswordResetTokenRepository`
    - `Create(ctx, userID, tokenDigest, expiresAt) error`
    - `DeleteUnusedByUserID(ctx, userID) error`
  - テストを追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 60 行（実装 30 行 + テスト 30 行）

- [x] **2-3**: [Go] UserRepository にユーザー作成・ユーザー名検索メソッドを追加、Profile/Setting/EmailNotification Repository を新規作成
  - `repository/user.go` に以下を追加:
    - `Create(ctx, params) (*model.User, error)` - ユーザー作成
    - `GetByUsername(ctx, username) error` - ユーザー名存在チェック（エラーのみ返す）
  - 新規ファイル `repository/profile.go`:
    - `ProfileRepository` 構造体、`NewProfileRepository`, `WithTx`, `Create(ctx, userID, name) error`
  - 新規ファイル `repository/setting.go`:
    - `SettingRepository` 構造体、`NewSettingRepository`, `WithTx`, `Create(ctx, userID) error`
  - 新規ファイル `repository/email_notification.go`:
    - `EmailNotificationRepository` 構造体、`NewEmailNotificationRepository`, `WithTx`, `Create(ctx, userID, unsubscriptionKey) error`
  - テストを追加
  - **想定ファイル数**: 約 8 ファイル（実装 4 + テスト 4）
  - **想定行数**: 約 250 行（実装 120 行 + テスト 130 行）

### フェーズ 3: UseCase層の修正

<!--
Repository拡張完了後、UseCase層のQuery直接依存を除去する。
テストファイルはdepguardルールから除外されているため、テストコードのquery import修正は任意。
ただし、コンストラクタ引数変更に伴うテストコード修正は必須。
-->

- [x] **3-1**: [Go] create_session.go を Repository経由に修正
  - `CreateSessionUsecase` のフィールドを `*query.Queries` → `*repository.SessionRepository` に変更
  - `NewCreateSessionUsecase` の引数を変更
  - `Execute` 内の `query.CreateSessionParams` 使用箇所を `SessionRepository.CreateSession` に委譲
  - `generatePrivateID` 関数を削除（SessionRepositoryに統一）
  - 呼び出し元の修正:
    - `usecase/complete_sign_up.go` 内の `NewCreateSessionUsecase` 呼び出し
    - `usecase/update_password_reset.go` 内の `NewCreateSessionUsecase` 呼び出し
    - `handler/sign_in_code/handler.go` のコンストラクタ引数
    - `handler/sign_in_password/handler.go` のコンストラクタ引数
    - `cmd/server/main.go` のUsecase初期化
  - テストの修正（コンストラクタ引数変更への対応）
  - **想定ファイル数**: 約 15 ファイル（実装 6 + テスト 9）
  - **想定行数**: 約 150 行（実装 60 行 + テスト 90 行）

- [x] **3-2**: [Go] create_password_reset_token.go を Repository経由に修正
  - `CreatePasswordResetTokenUsecase` のフィールドを `*query.Queries` → `*repository.PasswordResetTokenRepository` + `*repository.UserRepository` に変更
  - `NewCreatePasswordResetTokenUsecase` の引数を変更
  - `Execute` 内の Query直接呼び出しを Repository メソッドに委譲
  - 呼び出し元の修正:
    - `handler/password_reset/handler.go` のコンストラクタ引数
    - `cmd/server/main.go` のUsecase初期化
  - テストの修正（コンストラクタ引数変更への対応）
  - **想定ファイル数**: 約 5 ファイル（実装 3 + テスト 2）
  - **想定行数**: 約 100 行（実装 50 行 + テスト 50 行）

- [x] **3-3**: [Go] complete_sign_up.go を Repository経由に修正
  - `CompleteSignUpUsecase` のフィールドを `*query.Queries` → 各Repository に変更
  - `CompleteSignUpResult.User` の型を `query.CreateUserRow` → `*model.User` に変更
  - `NewCompleteSignUpUsecase` の引数を変更
  - `Execute` 内の Query直接呼び出しを Repository メソッドに委譲
  - 呼び出し元の修正:
    - `handler/sign_up_username/handler.go` のコンストラクタ引数
    - `handler/sign_up_username/create.go` の `ucResult.User` アクセス（型変更への対応）
    - `cmd/server/main.go` のUsecase初期化
  - テストの修正（コンストラクタ引数変更・Result型変更への対応）
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

### フェーズ 4: 検証

- [x] **4-1**: [Go] Lint・ビルド・テスト全体の通過確認
  - `make lint` で depguard 違反が0件であることを確認
  - `go build ./...` でビルドが通ることを確認
  - `make test` で全テストがパスすることを確認
  - **想定ファイル数**: 約 0 ファイル（実装 0 + テスト 0）
  - **想定行数**: 約 0 行（実装 0 行 + テスト 0 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **既存RepositoryメソッドのQuery戻り値型をModel型に統一**: 現在 `UserRepository.GetByEmailForSignIn` 等が `query.GetUserByEmailForSignInRow` を返しているが、これらの修正は影響範囲が広いため今回のスコープ外とする
- **テストコード内のquery import修正**: depguardルールでテストファイルは除外されているため、テストコード内で `query` を直接importしている箇所の修正は必須ではない（コンストラクタ引数変更に伴う修正のみ実施）

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [golangci-lint depguard設定](/workspace/go/.golangci.yml)
- [アーキテクチャガイド](/workspace/go/docs/architecture-guide.md)
