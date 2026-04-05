# UseCase オーケストレーション化リファクタリング 作業計画書

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

- [アーキテクチャガイド](/workspace/go/docs/architecture-guide.md)
- [ハンドラーガイド](/workspace/go/docs/handler-guide.md)
- [バリデーションガイド](/workspace/go/docs/validation-guide.md)

## 概要

<!--
ガイドライン:
- この機能が「何であるか」「なぜ必要か」を簡潔に説明
- 2-3段落程度で簡潔に
- 既存機能の変更の場合は、変更の背景と目的を記述
-->

Go 版 Annict の依存関係ルールを Wikino と統一するリファクタリングです。Handler を薄い Adapter に変え、UseCase をすべてのデータアクセスのオーケストレーターとして位置づけます。

**現状の課題**:

- Handler が Repository に直接依存しており、データアクセスの経路が UseCase 経由と Repository 直接の 2 通り存在する
- Validator が Handler パッケージ内（Presentation 層）に配置されており、UseCase からの呼び出しではなく Handler が直接呼び出している
- Wikino・Mewst と依存関係ルールが異なり、プロジェクト間の一貫性がない

**目標**:

- Handler のすべてのデータアクセスを UseCase 経由に統一する
- Validator を Application 層に移動し、UseCase から呼び出すパターンに統一する
- depguard ルールで Handler → Repository の直接依存を禁止する
- Worker を Presentation 層（薄い Adapter）として位置づける

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

- `make lint` が更新後の depguard ルールで違反なく成功すること
- `go build ./...` でビルドが通ること
- `make test` で既存テストがすべてパスすること
- 既存機能の動作に変更がないこと（リファクタリングのみ）

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

- **保守性**: リファクタリング後のコードが Wikino と同じアーキテクチャパターンに準拠していること
- **一貫性**: Annict・Wikino・Mewst の 3 プロジェクトで同一の依存関係ルールが適用されること

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
- [@go/docs/handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTPハンドラーガイドライン（**ファイル名は標準の9種類のみ**）
- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - バリデーションガイド

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

### 目標とするアーキテクチャ（Wikino と統一）

```
┌─────────────────────────────────────────────────────────┐
│ Presentation層                                          │
│ - Handler, Worker, ViewModel, Template, Middleware    │
└─────────────────────────────────────────────────────────┘
         ↓ 依存
┌─────────────────────────────────────────────────────────┐
│ Application層                                           │
│ - UseCase, Validator                                  │
└─────────────────────────────────────────────────────────┘
         ↓ 依存
┌─────────────────────────────────────────────────────────┐
│ Domain/Infrastructure層（統合）                          │
│ - Query (sqlc), Repository, Model, Dispatcher         │
└─────────────────────────────────────────────────────────┘
```

### 現在と目標の依存関係ルール比較

| ルール               | 現在 (Annict)                                         | 目標 (Wikino と統一)                               |
| -------------------- | ----------------------------------------------------- | -------------------------------------------------- |
| Handler → Repository | **許可**                                              | **禁止**（UseCase 経由のみ）                       |
| Handler → UseCase    | 許可                                                  | 許可                                               |
| Handler → Validator  | **許可**（Handler 内の validator.go）                 | **禁止**（UseCase 経由のみ）                       |
| UseCase → Validator  | 呼び出さない                                          | **UseCase が呼び出す**                             |
| UseCase → Repository | 許可                                                  | 許可                                               |
| UseCase → Dispatcher | **UseCase が `worker.Client` に直接依存**             | **UseCase が `dispatcher` 経由で投入**             |
| Dispatcher の配置    | なし（`worker.Client` に統合）                        | `internal/dispatcher/`（Domain/Infrastructure 層） |
| Validator の配置     | `internal/handler/**/validator.go`（Presentation 層） | `internal/validator/`（Application 層）            |
| Worker の層          | Application 層                                        | **Presentation 層**（薄い Adapter）                |

### 設計原則の変更

現在の設計原則:

- Handler は読み取り処理で Repository に直接アクセスできる
- UseCase はトランザクションを伴う書き込み処理のみに使用する
- Validator は Handler パッケージ内に配置し、Handler が直接呼び出す

目標の設計原則（Wikino と統一）:

- **Handler から Repository への直接依存は禁止**: Handler のすべてのデータアクセスは UseCase を経由する
- **UseCase はオーケストレーター**: 書き込み UseCase はバリデーション・ビジネスロジック・永続化を統括する。読み取り UseCase はデータ取得を担当する
- **Handler / Worker は薄い Adapter**: HTTP/ジョブの入出力変換のみ。バリデーションは UseCase を経由する
- **Validator は Application 層**: すべてのバリデーションは `internal/validator/` に配置し、UseCase から呼び出される
- **Dispatcher は Domain/Infrastructure 層**: ジョブキューへの投入は `internal/dispatcher/` に配置し、UseCase から呼び出される。UseCase が `worker` パッケージに直接依存することは禁止
- **Query への依存は Repository のみ**: Handler/UseCase/Worker が Query に直接依存することは禁止

### Handler の処理フロー変更

**現在のフロー**:

```
Handler:
  1. HTTP リクエストをパース
  2. Validator を呼び出し（Handler が直接）
  3. エラーがあればフォーム再表示
  4. UseCase を呼び出し（書き込みの場合のみ）
  5. または Repository を直接呼び出し（読み取りの場合）
  6. ViewModel に変換してレンダリング
```

**目標のフロー**:

```
Handler:
  1. HTTP リクエストをパース
  2. UseCase を呼び出し（読み取り・書き込み共通）
  3. UseCase の結果にバリデーションエラーがあればフォーム再表示
  4. ViewModel に変換してレンダリング

UseCase（書き込み）:
  1. Validator を呼び出し
  2. バリデーションエラーがあれば結果に含めて返却
  3. ビジネスロジック実行
  4. Repository 経由で永続化

UseCase（読み取り）:
  1. Repository 経由でデータ取得
  2. 結果を返却
```

### Handler → UseCase → Validator の実装パターン

```go
// internal/validator/sign_in.go（Application 層）
type CreateSignInValidator struct{}

type CreateSignInValidatorInput struct {
    Email string
}

type CreateSignInValidatorResult struct {
    FormErrors *session.FormErrors
}

func (v *CreateSignInValidator) Validate(ctx context.Context, input CreateSignInValidatorInput) *CreateSignInValidatorResult {
    formErrors := session.NewFormErrors()
    if input.Email == "" {
        formErrors.AddFieldError("email", i18n.T(ctx, "error_required"))
    }
    return &CreateSignInValidatorResult{FormErrors: formErrors}
}
```

```go
// internal/usecase/send_sign_in_code.go（Application 層）
type SendSignInCodeUsecase struct {
    validator *validator.CreateSignInValidator
    userRepo  *repository.UserRepository
    // ...
}

type SendSignInCodeOutput struct {
    FormErrors *session.FormErrors // バリデーションエラー（nilなら成功）
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

    return &SendSignInCodeOutput{}, nil
}
```

```go
// internal/handler/sign_in/create.go（Presentation 層）
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    result, err := h.sendSignInCodeUC.Execute(ctx, usecase.SendSignInCodeInput{
        Email: r.FormValue("email"),
    })
    if err != nil {
        // システムエラー
        return
    }
    if result.FormErrors != nil && result.FormErrors.HasErrors() {
        // フォーム再表示
        return
    }

    // 成功時のリダイレクト
}
```

### 読み取り UseCase の実装パターン

```go
// internal/usecase/get_popular_works.go
type GetPopularWorksUsecase struct {
    workRepo *repository.WorkRepository
}

type GetPopularWorksOutput struct {
    Works []*model.Work
}

func (uc *GetPopularWorksUsecase) Execute(ctx context.Context) (*GetPopularWorksOutput, error) {
    works, err := uc.workRepo.GetPopularWorksWithDetails(ctx)
    if err != nil {
        return nil, fmt.Errorf("人気作品の取得に失敗: %w", err)
    }
    return &GetPopularWorksOutput{Works: works}, nil
}
```

### Dispatcher パッケージの設計

現在、UseCase は `worker.Client`（`internal/worker/client.go`）に直接依存してジョブを投入している。Worker を Presentation 層に変更するため、ジョブ投入の責務を `internal/dispatcher/` パッケージ（Domain/Infrastructure 層）に分離する。

**現在の依存関係**:

```
UseCase → worker.Client → river.Client.Insert()
```

**目標の依存関係**:

```
UseCase → dispatcher.Dispatcher → river.Client.Insert()
Worker（Presentation層）→ ジョブの受信・処理のみ
```

**影響を受ける UseCase（3 ファイル）**:

| UseCase                          | 現在の依存      | 変更後の依存            |
| -------------------------------- | --------------- | ----------------------- |
| `send_sign_in_code.go`           | `worker.Client` | `dispatcher.Dispatcher` |
| `send_sign_up_code.go`           | `worker.Client` | `dispatcher.Dispatcher` |
| `create_password_reset_token.go` | `worker.Client` | `dispatcher.Dispatcher` |

**Dispatcher の実装パターン**:

```go
// internal/dispatcher/dispatcher.go
package dispatcher

import (
    "context"

    "github.com/jackc/pgx/v5"
    "github.com/riverqueue/river"

    "github.com/annict/annict/go/internal/worker"
)

// Dispatcher はバックグラウンドジョブの投入を担当する
type Dispatcher struct {
    riverClient *river.Client[pgx.Tx]
}

// NewDispatcher は新しい Dispatcher を作成する
func NewDispatcher(riverClient *river.Client[pgx.Tx]) *Dispatcher {
    return &Dispatcher{riverClient: riverClient}
}

// InsertSendEmail はメール送信ジョブを投入する
func (d *Dispatcher) InsertSendEmail(ctx context.Context, args worker.SendEmailArgs) error {
    _, err := d.riverClient.Insert(ctx, args, nil)
    return err
}
```

**Worker パッケージの変更**:

- `worker/client.go` の `Client` 構造体からジョブ投入メソッドを除去（ライフサイクル管理のみに限定）
- ジョブ引数の型定義（`SendEmailArgs` 等）は Worker パッケージに残す（River のワーカー登録で必要なため）
- Dispatcher は Worker パッケージのジョブ引数型を import する（Domain/Infrastructure 層から Presentation 層への依存は型定義のみ許可）

### Validator ファイルの命名規則（`internal/validator/` に移動後）

現在の `internal/handler/{resource}/validator.go` を `internal/validator/` に移動する際の命名規則:

| 現在のファイル                             | 移動先                             | 構造体名                            |
| ------------------------------------------ | ---------------------------------- | ----------------------------------- |
| `handler/sign_in/validator.go`             | `validator/sign_in.go`             | `CreateSignInValidator`             |
| `handler/sign_in_code/validator.go`        | `validator/sign_in_code.go`        | `CreateSignInCodeValidator`         |
| `handler/sign_in_password/validator.go`    | `validator/sign_in_password.go`    | `CreateSignInPasswordValidator`     |
| `handler/sign_up/validator.go`             | `validator/sign_up.go`             | `CreateSignUpValidator`             |
| `handler/sign_up_code/validator.go`        | `validator/sign_up_code.go`        | `CreateSignUpCodeValidator`         |
| `handler/sign_up_username/validator.go`    | `validator/sign_up_username.go`    | `CreateSignUpUsernameValidator`     |
| `handler/password_reset/validator.go`      | `validator/password_reset.go`      | `CreatePasswordResetValidator`      |
| `handler/password/validator.go`            | `validator/password.go`            | `UpdatePasswordValidator`           |
| `handler/db_work/validator.go`             | `validator/db_work.go`             | `CreateDbWorkValidator`             |
| `handler/supporters_checkout/validator.go` | `validator/supporters_checkout.go` | `CreateSupportersCheckoutValidator` |

### 現在の Handler → Repository 依存の一覧と対応方針

#### 読み取り専用（新規 UseCase 作成が必要）

| Handler                 | Repository メソッド                                                         | 作成する UseCase            |
| ----------------------- | --------------------------------------------------------------------------- | --------------------------- |
| `popular_work/index.go` | `WorkRepository.GetPopularWorksWithDetails`                                 | `GetPopularWorksUsecase`    |
| `health/show.go`        | `WorkRepository.GetByID`                                                    | `CheckHealthUsecase`        |
| `ics/show.go`           | `UserCalendarRepository.GetByUsername`                                      | `GetUserCalendarUsecase`    |
| `supporters/show.go`    | `StripeSubscriberRepository.GetByID`, `GumroadSubscriberRepository.GetByID` | `GetSupporterStatusUsecase` |
| `db_work/index.go`      | `WorkRepository.ListForDB`, `WorkRepository.CountForDB`                     | `ListDbWorksUsecase`        |

#### 書き込み Handler の前処理読み取り（既存 UseCase の拡張またはフロー変更）

| Handler                         | Repository メソッド                        | 対応方針                                  |
| ------------------------------- | ------------------------------------------ | ----------------------------------------- |
| `sign_in/create.go`             | `UserRepository.GetByEmailForSignIn`       | `SendSignInCodeUsecase` に統合            |
| `sign_in_code/create.go`        | `UserRepository.GetByID`                   | `VerifySignInCodeUsecase` の出力に含める  |
| `sign_in_password/create.go`    | `UserRepository.GetByEmailOrUsername`      | 新規 `AuthenticateByPasswordUsecase` 作成 |
| `sign_up/create.go`             | `UserRepository.GetByEmail`                | `SendSignUpCodeUsecase` に統合            |
| `password/edit.go`              | `PasswordResetTokenRepository.GetByDigest` | 新規 `GetPasswordResetTokenUsecase` 作成  |
| `supporters_checkout/create.go` | `StripeSubscriberRepository.GetByID`       | 新規 `CreateCheckoutSessionUsecase` 作成  |
| `supporters_portal/create.go`   | `StripeSubscriberRepository.GetByID`       | 新規 `CreatePortalSessionUsecase` 作成    |
| `webhooks/stripe/create.go`     | `StripeWebhookEventRepository.*`           | 新規 `ProcessStripeWebhookUsecase` 作成   |

### depguard ルール変更

#### Handler 層（追加）

```yaml
handler-layer:
  files:
    - "**/internal/handler/**"
  deny:
    - pkg: github.com/annict/annict/go/internal/query
    - pkg: github.com/annict/annict/go/internal/repository # 追加
    - pkg: github.com/annict/annict/go/internal/validator # 追加
```

#### Validator 層（変更: Handler 内 → Application 層）

```yaml
# 旧ルール（削除）
validator-layer:
  files:
    - "**/internal/handler/**/validator.go"
  deny:
    - pkg: github.com/annict/annict/go/internal/query
    - pkg: github.com/annict/annict/go/internal/usecase

# 新ルール（追加）
validator-layer:
  files:
    - "**/internal/validator/*.go"
  deny:
    - pkg: github.com/annict/annict/go/internal/query
    - pkg: github.com/annict/annict/go/internal/handler
    - pkg: github.com/annict/annict/go/internal/middleware
    - pkg: github.com/annict/annict/go/internal/viewmodel
    - pkg: github.com/annict/annict/go/internal/templates
    - pkg: github.com/annict/annict/go/internal/usecase
```

#### Application 層（変更: Worker への依存禁止を追加）

```yaml
application-layer:
  files:
    - "**/internal/usecase/**"
  deny:
    - pkg: github.com/annict/annict/go/internal/query
    - pkg: github.com/annict/annict/go/internal/handler
    - pkg: github.com/annict/annict/go/internal/middleware
    - pkg: github.com/annict/annict/go/internal/viewmodel
    - pkg: github.com/annict/annict/go/internal/templates
    - pkg: github.com/annict/annict/go/internal/worker # 追加（Dispatcher 経由で投入する）
```

#### Worker 層（変更: Application → Presentation）

```yaml
worker-layer:
  files:
    - "**/internal/worker/*.go"
  deny:
    - pkg: github.com/annict/annict/go/internal/query
    - pkg: github.com/annict/annict/go/internal/repository # 追加（Presentation層なので）
    - pkg: github.com/annict/annict/go/internal/handler
    - pkg: github.com/annict/annict/go/internal/middleware
    - pkg: github.com/annict/annict/go/internal/viewmodel
```

#### Dispatcher 層（新規: Domain/Infrastructure 層）

```yaml
dispatcher-layer:
  files:
    - "**/internal/dispatcher/*.go"
  deny:
    - pkg: github.com/annict/annict/go/internal/usecase
    - pkg: github.com/annict/annict/go/internal/handler
    - pkg: github.com/annict/annict/go/internal/middleware
    - pkg: github.com/annict/annict/go/internal/viewmodel
    - pkg: github.com/annict/annict/go/internal/templates
    - pkg: github.com/annict/annict/go/internal/query
```

## 採用しなかった方針

<!--
ガイドライン:
- 検討したが採用しなかった設計や機能を、理由とともに記述
- 将来の開発者が同じ検討を繰り返さないための判断記録
- タスク完了後、この内容は `specs/` の仕様書にも転記する
- 該当がない場合は「なし」と記載
-->

### Handler から Repository への依存を引き続き許可する

読み取り専用の処理は Handler から直接 Repository を呼び出す方が記述量が少なく簡潔になるが、以下の理由で不採用とした:

- Wikino・Mewst との一貫性が損なわれる
- データアクセスの経路が UseCase 経由と Repository 直接の 2 通りになり、「どちらを使うか」の判断コストが発生する
- UseCase に統一することで、将来的にキャッシュ層やロギングの追加が容易になる

### Policy パッケージの新規作成

Wikino では `internal/policy/` パッケージ（Domain/Infrastructure 層）が存在し、UseCase から認可チェックを呼び出すパターンを採用している。しかし、Annict では現時点で認可ロジック（リソース所有者チェック等）が限定的であるため、今回のスコープでは Policy パッケージの作成は行わない。認可ロジックが増えた時点で別途作業計画書を作成する。

### Validator を UseCase から分離して Handler で直接呼び出し続ける

現在のパターン（Handler → Validator → UseCase）はシンプルで理解しやすいが、以下の理由で不採用とした:

- Wikino・Mewst では UseCase がバリデーションを統括するパターンに統一済み
- UseCase がオーケストレーターとして一貫した処理フローを提供することで、Handler の責務が HTTP 入出力のみに限定される
- バリデーションを UseCase に含めることで、Handler 以外のエントリポイント（Worker 等）からも同じバリデーションロジックを再利用できる

### sessions テーブルに flash メッセージを格納し続ける

現在の Annict は Rails の ActiveRecord SessionStore 互換で、flash メッセージを `sessions` テーブルの `data` カラムに JSON として格納している。このため `CreateSessionUsecase` がセッション作成時に `flashMessage` パラメータを受け取り、UseCase に Presentation 層の関心事（flash メッセージ）が漏れている。

Wikino では flash メッセージを Cookie ベースで管理しており（`session.FlashManager`）、UseCase はセッション作成のみに専念している。Go 版のページで設定した flash は Go 版のページでのみ表示されるため、Rails 版との互換性は不要。よって Wikino と同様に Cookie ベースの flash 管理に移行する（フェーズ 2b）。

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

### フェーズ 1: depguard ルール整備

<!--
depguard ルールの追加・修正を行う。この時点では Handler → Repository 禁止は有効化しない（段階的に移行するため）。
Worker 層ルールの追加と glob パターン修正は docs/plans/2_todo/golangci-lint-worker-layer.md の内容を含む。
-->

- [x] **1-1**: [Go] Worker 層の depguard ルール追加と glob パターン修正
  - `.golangci.yml` に worker-layer ルールを追加（Worker は Presentation 層として、Query・Repository・Handler・Middleware・ViewModel への依存を禁止）
  - viewmodel, middleware, model, repository の glob パターンを `**` → `*.go` に修正
  - Worker が既存で Query に直接依存していたため、以下のリファクタリングも実施:
    - クリーンアップロジックを UseCase に抽出（`cleanup_expired_tokens.go`, `cleanup_expired_sign_in_codes.go`）
    - Worker をインターフェース経由で UseCase に依存する薄い Adapter に変更
    - Repository に `DeleteExpired` メソッドを追加
  - `make lint` で違反なし確認済み
  - **実績ファイル数**: 12 ファイル（実装 8 + テスト 2 + 設定 1 + 計画書 1）
  - **実績行数**: 約 300 行（実装 215 行 + テスト 82 行）

- [x] **1-2**: [Go] Validator（Application 層）と Dispatcher（Domain/Infrastructure 層）の depguard ルールを追加
  - `.golangci.yml` に新しい validator-layer ルールを追加（`**/internal/validator/*.go` に対して、Handler・Middleware・ViewModel・Templates・Query・UseCase への依存を禁止）
  - `.golangci.yml` に新しい dispatcher-layer ルールを追加（`**/internal/dispatcher/*.go` に対して、UseCase・Handler・Middleware・ViewModel・Templates・Query への依存を禁止）
  - `.golangci.yml` の application-layer ルールに `internal/worker` への依存禁止を追加（Dispatcher 経由に統一するため）。ただし、UseCase → Worker の禁止はフェーズ 2a 完了後に有効化する
  - 旧 validator-layer ルール（`**/internal/handler/**/validator.go`）はフェーズ 2 完了後に削除するため、この時点では両方のルールが共存する
  - `make lint` で違反がないことを確認
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

### フェーズ 2a: Dispatcher パッケージの作成と UseCase の Worker 依存除去

<!--
Worker を Presentation 層に変更する前提として、UseCase が worker パッケージに依存している箇所を
Dispatcher 経由に変更する。
-->

- [x] **2a-1**: [Go] Dispatcher パッケージの作成と UseCase の移行
  - `internal/dispatcher/dispatcher.go` を新規作成（`Dispatcher` 構造体、`NewDispatcher`、`InsertSendEmail` メソッド）
  - `usecase/send_sign_in_code.go` の `worker.Client` 依存を `dispatcher.Dispatcher` に変更
  - `usecase/send_sign_up_code.go` の `worker.Client` 依存を `dispatcher.Dispatcher` に変更
  - `usecase/create_password_reset_token.go` の `worker.Client` 依存を `dispatcher.Dispatcher` に変更
  - `cmd/server/main.go` で `Dispatcher` を初期化し、UseCase に注入
  - テスト修正（コンストラクタ引数変更への対応）
  - `.golangci.yml` の application-layer ルールに `internal/worker` への依存禁止を有効化
  - **想定ファイル数**: 約 8 ファイル（実装 5 + テスト 3）
  - **想定行数**: 約 200 行（実装 100 行 + テスト 100 行）

### フェーズ 2: 認証系の Validator 移動と UseCase オーケストレーション化

<!--
認証系（sign_in, sign_up, password 等）の Validator を Application 層に移動し、
UseCase がバリデーションを統括するパターンに変更する。
各タスクでは Validator 移動 + UseCase 修正 + Handler 修正を1セットで行う。
-->

- [x] **2-1**: [Go] sign_in の Validator 移動と UseCase 修正
  - `handler/sign_in/validator.go` → `validator/sign_in.go` に移動
  - `SendSignInCodeUsecase` に Validator 呼び出しを追加し、`UserRepository.GetByEmailForSignIn` の呼び出しも UseCase に移動
  - `handler/sign_in/create.go` から Validator・Repository の直接呼び出しを除去し、UseCase の結果に基づいてエラー処理
  - `handler/sign_in/handler.go` から Repository・Validator のフィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 8 ファイル（実装 5 + テスト 3）
  - **想定行数**: 約 200 行（実装 120 行 + テスト 80 行）

- [x] **2-2**: [Go] sign_in_password の Validator 移動と UseCase 作成
  - `handler/sign_in_password/validator.go` → `validator/sign_in_password.go` に移動
  - 新規 `AuthenticateByPasswordUsecase` を作成（Validator 呼び出し + `UserRepository.GetByEmailOrUsername` + パスワード照合 + セッション作成）
  - `handler/sign_in_password/create.go` から Validator・Repository の直接呼び出しを除去
  - `handler/sign_in_password/handler.go` から Repository・Validator のフィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 8 ファイル（実装 5 + テスト 3）
  - **想定行数**: 約 250 行（実装 150 行 + テスト 100 行）

- [x] **2-3**: [Go] sign_in_code の Validator 移動と UseCase 修正
  - `handler/sign_in_code/validator.go` → `validator/sign_in_code.go` に移動
  - `VerifySignInCodeUsecase` に Validator 呼び出しを追加し、`UserRepository.GetByID` の呼び出しも UseCase に移動
  - `handler/sign_in_code/create.go` から Validator・Repository の直接呼び出しを除去
  - `handler/sign_in_code/handler.go` から Repository・Validator のフィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 8 ファイル（実装 5 + テスト 3）
  - **想定行数**: 約 200 行（実装 120 行 + テスト 80 行）

- [x] **2-4**: [Go] sign_up の Validator 移動と UseCase 修正
  - `handler/sign_up/validator.go` → `validator/sign_up.go` に移動
  - `SendSignUpCodeUsecase` に Validator 呼び出しを追加し、`UserRepository.GetByEmail` の呼び出しも UseCase に移動
  - `handler/sign_up/create.go` から Validator・Repository の直接呼び出しを除去
  - `handler/sign_up/handler.go` から Repository・Validator のフィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 8 ファイル（実装 5 + テスト 3）
  - **想定行数**: 約 200 行（実装 120 行 + テスト 80 行）

- [x] **2-5**: [Go] sign_up_code の Validator 移動と UseCase 修正
  - `handler/sign_up_code/validator.go` → `validator/sign_up_code.go` に移動
  - `VerifySignUpCodeUsecase` に Validator 呼び出しを追加
  - `handler/sign_up_code/create.go` から Validator の直接呼び出しを除去
  - テスト修正
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

- [x] **2-6**: [Go] sign_up_username の Validator 移動と UseCase 修正
  - `handler/sign_up_username/validator.go` → `validator/sign_up_username.go` に移動
  - `CompleteSignUpUsecase` に Validator 呼び出しを追加
  - `handler/sign_up_username/create.go` から Validator の直接呼び出しを除去
  - テスト修正
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

### フェーズ 2b: Flash メッセージの Cookie ベース化

<!--
Wikino と同様に flash メッセージを Cookie ベースで管理するように変更する。
これにより CreateSessionUsecase から flashMessage パラメータを除去し、
flash メッセージを完全に Presentation 層の関心事に閉じ込める。
-->

- [x] **2b-1**: [Go] FlashManager の導入と CreateSessionUsecase からの flash 除去
  - Wikino の `FlashManager`（`/wikino/go/internal/session/flash.go`）を参考に、Cookie ベースの flash メッセージ管理を導入
  - `internal/session/flash_manager.go` を新規作成（`FlashManager` 構造体、`SetSuccess`/`SetError`/`GetFlash` メソッド）
  - `CreateSessionUsecase.Execute` から `flashMessage` パラメータを除去
  - `AuthenticateByPasswordUsecase.Execute` から `flashJSON` パラメータを除去
  - 全 Handler の sign-in/sign-up フローで、UseCase 成功後に `FlashManager.SetSuccess()` を呼び出すパターンに変更
  - flash メッセージの読み取り側（テンプレートのレイアウト等）を Cookie ベースに変更
  - テスト修正
  - **想定ファイル数**: 約 12 ファイル（実装 8 + テスト 4）
  - **想定行数**: 約 300 行（実装 200 行 + テスト 100 行）

### フェーズ 3: パスワード系の Validator 移動と UseCase 修正

- [x] **3-1**: [Go] password_reset の Validator 移動と UseCase 修正
  - `handler/password_reset/validator.go` → `validator/password_reset.go` に移動
  - `CreatePasswordResetTokenUsecase` に Validator 呼び出しを追加
  - `handler/password_reset/create.go` から Validator の直接呼び出しを除去
  - `handler/password_reset/handler.go` から Validator・Repository のフィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

- [x] **3-2**: [Go] password の Validator 移動と UseCase 修正
  - `handler/password/validator.go` → `validator/password.go` に移動
  - `UpdatePasswordResetUsecase` に Validator 呼び出しを追加し、`PasswordResetTokenRepository.GetByDigest` の呼び出しも UseCase に移動
  - 新規 `GetPasswordResetTokenUsecase`（edit.go 用の読み取り UseCase）を作成
  - `handler/password/edit.go` と `handler/password/create.go` から Validator・Repository の直接呼び出しを除去
  - `handler/password/handler.go` から Repository・Validator のフィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 10 ファイル（実装 6 + テスト 4）
  - **想定行数**: 約 250 行（実装 150 行 + テスト 100 行）

### フェーズ 4: 読み取り専用 Handler の UseCase 化

<!--
Repository を直接呼び出している読み取り専用の Handler に対して、読み取り UseCase を作成する。
-->

- [x] **4-1**: [Go] popular_work の読み取り UseCase 作成
  - 新規 `GetPopularWorksUsecase` を作成（`WorkRepository.GetPopularWorksWithDetails` を呼び出す）
  - `handler/popular_work/index.go` を UseCase 経由に変更
  - `handler/popular_work/handler.go` から Repository フィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 100 行（実装 60 行 + テスト 40 行）

- [x] **4-2**: [Go] health の読み取り UseCase 作成
  - 新規 `CheckHealthUsecase` を作成（`WorkRepository.GetByID` を呼び出す DB ヘルスチェック）
  - `handler/health/show.go` を UseCase 経由に変更
  - `handler/health/handler.go` から Repository フィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 80 行（実装 50 行 + テスト 30 行）

- [x] **4-3**: [Go] ics の読み取り UseCase 作成
  - 新規 `GetUserCalendarUsecase` を作成（`UserCalendarRepository.GetByUsername` を呼び出す）
  - `handler/ics/show.go` を UseCase 経由に変更
  - `handler/ics/handler.go` から Repository フィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 100 行（実装 60 行 + テスト 40 行）

- [x] **4-4**: [Go] supporters の読み取り UseCase 作成
  - 新規 `GetSupporterStatusUsecase` を作成（`StripeSubscriberRepository` + `GumroadSubscriberRepository` を呼び出す）
  - `handler/supporters/show.go` を UseCase 経由に変更
  - `handler/supporters/handler.go` から Repository フィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 150 行（実装 90 行 + テスト 60 行）

- [x] **4-5**: [Go] db_work の Validator 移動と読み取り UseCase 作成
  - `handler/db_work/validator.go` → `validator/db_work.go` に移動
  - 新規 `ListDbWorksUsecase` を作成（`WorkRepository.ListForDB` + `WorkRepository.CountForDB` を呼び出す）
  - 既存の `CreateWorkUsecase` に Validator 呼び出しを追加
  - `handler/db_work/index.go` と `handler/db_work/create.go` から Repository・Validator の直接呼び出しを除去
  - `handler/db_work/handler.go` から Repository・Validator のフィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 10 ファイル（実装 6 + テスト 4）
  - **想定行数**: 約 250 行（実装 150 行 + テスト 100 行）

### フェーズ 5: Stripe 関連 Handler の UseCase 化

- [x] **5-1**: [Go] supporters_checkout の Validator 移動と UseCase 作成
  - `handler/supporters_checkout/validator.go` → `validator/supporters_checkout.go` に移動
  - 新規 `CreateCheckoutSessionUsecase` を作成（`StripeSubscriberRepository` の読み取り + Stripe API 呼び出し + Validator 統合）
  - `handler/supporters_checkout/create.go` から Repository・Validator の直接呼び出しを除去
  - `handler/supporters_checkout/handler.go` から Repository・Validator のフィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 8 ファイル（実装 5 + テスト 3）
  - **想定行数**: 約 200 行（実装 120 行 + テスト 80 行）

- [x] **5-2**: [Go] supporters_portal の UseCase 作成
  - 新規 `CreatePortalSessionUsecase` を作成（`StripeSubscriberRepository` の読み取り + Stripe API 呼び出し）
  - `handler/supporters_portal/create.go` から Repository の直接呼び出しを除去
  - `handler/supporters_portal/handler.go` から Repository フィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 150 行（実装 90 行 + テスト 60 行）

- [x] **5-3**: [Go] webhooks/stripe の UseCase 作成
  - 新規 `ProcessStripeWebhookUsecase` を作成（`StripeWebhookEventRepository` の冪等性チェック・イベント記録 + 既存 UseCase の呼び出し）
  - `handler/webhooks/stripe/create.go` から Repository の直接呼び出しを除去
  - `handler/webhooks/stripe/handler.go` から Repository フィールドを除去
  - テスト修正
  - **想定ファイル数**: 約 8 ファイル（実装 5 + テスト 3）
  - **想定行数**: 約 250 行（実装 150 行 + テスト 100 行）

### フェーズ 6: depguard の Handler → Repository 禁止ルール有効化

<!--
すべての Handler から Repository 依存が除去された後に、depguard ルールを有効化する。
-->

- [x] **6-1**: [Go] depguard で Handler → Repository の依存を禁止
  - `.golangci.yml` の handler-layer ルールに `internal/repository` と `internal/validator` への依存禁止を追加
  - 旧 validator-layer ルール（`**/internal/handler/**/validator.go`）を削除
  - `make lint` で違反が 0 件であることを確認
  - `go build ./...` と `make test` でビルド・テストが通ることを確認
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 15 行（実装 15 行 + テスト 0 行）

### フェーズ 7: ドキュメント更新

- [x] **7-1**: [Go] CLAUDE.md とアーキテクチャガイドの更新
  - `go/CLAUDE.md` の「重要な設計原則」セクションを Wikino と統一
  - `go/CLAUDE.md` の「UsecaseとRepositoryの使い分け」セクションを更新（読み取り UseCase の追加）
  - `go/CLAUDE.md` の「バリデーション」セクションを更新（Validator の配置変更）
  - `go/docs/architecture-guide.md` のレイヤー図・依存関係ルール・Worker の位置づけを更新
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 200 行（実装 200 行 + テスト 0 行）

- [x] **7-2**: [Go] ハンドラーガイドとバリデーションガイドの更新
  - `go/docs/handler-guide.md` の Validator 配置セクションを更新（`validator.go` はもう handler 内に置かない）
  - `go/docs/validation-guide.md` を全面改訂（Validator の配置が `internal/validator/` に変更）
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 200 行（実装 200 行 + テスト 0 行）

### フェーズ 8: 仕様書への反映

<!--
**重要**: 実装完了後、必ず仕様書を作成・更新してください。
- 新しい機能の場合: `docs/specs/` に仕様書を新規作成する
- 既存機能の変更の場合: 対応する仕様書を最新の状態に更新する
- 概要・仕様・設計・採用しなかった方針を作業計画書から転記・整理する
-->

- [ ] **8-1**: 仕様書の作成・更新
  - `docs/specs/` に仕様書を作成または更新する
  - 作業計画書の概要・要件・設計・採用しなかった方針を仕様書に反映する

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **Policy パッケージの作成**: 現時点で認可ロジックが限定的なため。認可ロジックが増えた時点で別途対応
- **既存 Repository メソッドの Query 戻り値型を Model 型に統一**: 影響範囲が広いため今回のスコープ外
- **テストコード内の Repository import 修正**: depguard ルールでテストファイルは除外されているため必須ではない（コンストラクタ引数変更に伴う修正のみ実施）

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Wikino 開発ガイド（Go 版）](/workspace/wikino/go/CLAUDE.md) - 目標とするアーキテクチャの参考
- [Wikino アーキテクチャガイド](/workspace/wikino/go/docs/architecture-guide.md) - 依存関係ルールの参考
- [depguard アーキテクチャ違反の修正 作業計画書](/workspace/docs/plans/3_done/202602/fix-depguard-architecture-violations.md) - 前回の depguard 修正
- [golangci-lint Worker 層ルール追加 設計書](/workspace/docs/plans/2_todo/golangci-lint-worker-layer.md) - Worker 層ルールの元設計書
- [バリデーション責務のリファクタリング 設計書](/workspace/docs/plans/3_done/202602/validation-responsibility-refactoring.md) - バリデーション設計方針の参考
