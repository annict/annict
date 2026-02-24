# Worker実装統一 設計書

## 実装ガイドラインの参照

### Go版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - 全体的なコーディング規約
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド

## 概要

Annict、Wikino、MewstのGo版におけるWorker（バックグラウンドジョブ処理）実装を調査し、各プロジェクトの良いところを抽出して統一した実装パターンを策定する。

**目的**:

- 3つのプロジェクト間でWorker実装を統一し、保守性を向上させる
- 各プロジェクトの良いパターンを他プロジェクトにも反映する
- 新機能追加時の実装コストを削減する

**背景**:

- 現在、3つのプロジェクトでWorker実装にばらつきがある
- データベースドライバー、ワーカー登録方法、ジョブ定義などが異なる
- 統一することで横展開が容易になる

## 現状分析

### 各プロジェクトの実装比較

| 項目 | Mewst | Annict | Wikino |
|------|-------|--------|--------|
| **DBドライバー** | `database/sql` + `riverdatabasesql` | `pgx/v5` + `riverpgxv5` | `pgx/v5` + `riverpgxv5` |
| **コネクションプール** | なし | あり（詳細設定） | あり（詳細設定） |
| **ワーカー登録** | main.goで外部登録 | client.go内部で登録 | client.go内部で登録 |
| **InsertOpts定義** | Argsにメソッド定義 | なし | なし |
| **テンプレートレンダリング** | 事前レンダリング | ジョブ内レンダリング | ジョブ内レンダリング |
| **ログ出力** | 最小限 | 詳細 | 詳細 |
| **定期ジョブ** | なし | あり（PeriodicJob） | なし |
| **エンキューメソッド** | 汎用Insert | Client()経由 | 専用メソッド |

### 各プロジェクトの良い点

**Mewst:**
- ジョブArgsに`InsertOpts()`メソッドを定義 → リトライ回数などの設定がジョブ定義に近い場所にある
- テンプレートを事前レンダリング → ジョブのシリアライズが単純、テンプレート変更時に再キュー不要

**Annict:**
- `pgx/v5`を使用 → パフォーマンスが良い
- コネクションプール設定が詳細 → 本番運用に適している
- 定期ジョブ（PeriodicJob）を実装 → クリーンアップなどの定期処理が可能
- ログ出力が詳細 → 運用・デバッグしやすい

**Wikino:**
- `EnqueueEmailConfirmation`のような専用メソッド → 使いやすいAPI
- email.Senderがtempl.Component対応 → テンプレートの型安全性

## 要件

### 機能要件

- 3つのプロジェクトで同じWorker実装パターンを使用する
- DBドライバーは`pgx/v5` + `riverpgxv5`に統一する
- コネクションプール設定を標準化する
- ジョブArgsに`InsertOpts()`メソッドを定義するパターンを採用する
- ログ出力を標準化する（開始、完了、エラー時のログ）
- 定期ジョブの仕組みを標準化する

### 非機能要件

- **保守性**: 3つのプロジェクト間でコードパターンが統一され、学習コストが低い
- **テスト容易性**: ワーカーのテストが容易に書ける構造

## 設計

### 統一パターン

#### 1. データベースドライバー

**採用**: `pgx/v5` + `riverpgxv5`

```go
import (
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/riverqueue/river"
    "github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type Client struct {
    riverClient *river.Client[pgx.Tx]
    pool        *pgxpool.Pool
}
```

**理由**:
- Annict/Wikinoで既に使用している
- `database/sql`より高パフォーマンス
- PostgreSQL固有の機能をフル活用できる

#### 2. コネクションプール設定

**採用**: 標準設定を定義

```go
poolConfig.MaxConns = 10
poolConfig.MinConns = 2
poolConfig.MaxConnLifetime = 5 * time.Minute
poolConfig.MaxConnIdleTime = 2 * time.Minute
```

#### 3. ワーカー登録方法

**採用**: client.go内部で登録（Annict/Wikino方式）

**理由**:
- ワーカー関連のコードが1箇所に集約される
- main.goがシンプルになる
- 依存関係の管理が明確

```go
// client.go
func NewClient(ctx context.Context, databaseURL string, cfg *config.Config, deps Dependencies) (*Client, error) {
    // ワーカーの登録
    workers := river.NewWorkers()
    river.AddWorker(workers, NewSendEmailWorker(deps.EmailSender))

    // ...
}
```

#### 4. ジョブArgs定義

**採用**: `InsertOpts()`メソッドを定義（Mewst方式）

**理由**:
- リトライ回数、キューなどの設定がジョブの定義に近い場所にある
- ジョブ固有の設定が明示的

```go
type SendEmailArgs struct {
    To       string `json:"to"`
    Subject  string `json:"subject"`
    HTMLBody string `json:"html_body"`
    TextBody string `json:"text_body"`
}

func (SendEmailArgs) Kind() string {
    return "send_email"
}

func (SendEmailArgs) InsertOpts() river.InsertOpts {
    return river.InsertOpts{
        Queue:       river.QueueDefault,
        MaxAttempts: 5,
    }
}
```

#### 5. テンプレートレンダリング

**採用**: 事前レンダリング（Mewst方式）

**理由**:
- ジョブのシリアライズが単純（文字列のみ）
- テンプレート変更時に既存ジョブの再キュー不要
- ジョブの再実行時にも同じ内容が送信される

```go
// ユースケースでレンダリング
htmlStr, textStr := renderEmailTemplates(ctx, htmlBody, textBody)

// ジョブにはレンダリング済み文字列を渡す
workerClient.Insert(ctx, worker.SendEmailArgs{
    To:       email,
    Subject:  subject,
    HTMLBody: htmlStr,
    TextBody: textStr,
})
```

#### 6. ログ出力

**採用**: 詳細ログ（Annict/Wikino方式）

```go
func (w *SendEmailWorker) Work(ctx context.Context, job *river.Job[SendEmailArgs]) error {
    slog.InfoContext(ctx, "メール送信ジョブを開始します",
        "to", job.Args.To,
        "subject", job.Args.Subject,
    )

    err := w.sender.SendRaw(ctx, email.SendRawInput{...})
    if err != nil {
        slog.ErrorContext(ctx, "メール送信に失敗しました",
            "to", job.Args.To,
            "error", err,
        )
        return fmt.Errorf("メール送信に失敗: %w", err)
    }

    slog.InfoContext(ctx, "メール送信が完了しました",
        "to", job.Args.To,
    )
    return nil
}
```

#### 7. 定期ジョブ

**採用**: PeriodicJob（Annict方式）を標準化

```go
// main.goで定期ジョブを登録
periodicJob := river.NewPeriodicJob(
    dailyAt2AMSchedule{},
    func() (river.JobArgs, *river.InsertOpts) {
        return worker.CleanupExpiredTokensArgs{}, nil
    },
    nil,
)
riverClient.Client().PeriodicJobs().Add(periodicJob)
```

#### 8. Clientインターフェース

**採用**: 統一されたClientインターフェース

```go
type Client struct {
    riverClient *river.Client[pgx.Tx]
    pool        *pgxpool.Pool
}

// Start はワーカーの処理を開始する
func (c *Client) Start(ctx context.Context) error

// Stop はワーカーの処理を停止する
func (c *Client) Stop(ctx context.Context) error

// Insert はジョブをキューに追加する
func (c *Client) Insert(ctx context.Context, args river.JobArgs) (*rivertype.JobInsertResult, error)

// Client は内部のRiverクライアントへのアクセスを提供する（定期ジョブ用）
func (c *Client) Client() *river.Client[pgx.Tx]
```

### 統一後のディレクトリ構造

```
internal/worker/
├── client.go              # Workerクライアント
├── send_email.go          # メール送信ジョブ
├── send_email_test.go     # テスト
├── cleanup_expired_*.go   # クリーンアップジョブ（必要に応じて）
└── cleanup_expired_*_test.go
```

## タスクリスト

### フェーズ 1: Mewst修正

- [x] **1-1**: [Go] MewstのWorkerをpgx/v5に移行

  - `database/sql`から`pgx/v5`に変更
  - コネクションプール設定を追加
  - ワーカー登録をclient.go内部に移動
  - **想定ファイル数**: 約 3 ファイル（実装 2 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

- [x] **1-2**: [Go] Mewstのジョブにログ出力を追加

  - SendEmailWorkerに開始・完了・エラーログを追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 30 行 + テスト 20 行）

### フェーズ 2: Annict修正

- [x] **2-1**: [Go] AnnictのジョブArgsにInsertOptsメソッドを追加

  - 各ジョブArgsに`InsertOpts()`メソッドを定義
  - リトライ回数などの設定を明示化
  - **想定ファイル数**: 約 6 ファイル（実装 5 + テスト 1）
  - **想定行数**: 約 80 行（実装 60 行 + テスト 20 行）

- [x] **2-2**: [Go] Annictのメール送信を事前レンダリング方式に変更

  - ユースケースでテンプレートをレンダリング
  - ジョブにはレンダリング済み文字列を渡す
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 200 行（実装 150 行 + テスト 50 行）

### フェーズ 3: Wikino修正

- [x] **3-1**: [Go] WikinoのジョブArgsにInsertOptsメソッドを追加

  - 各ジョブArgsに`InsertOpts()`メソッドを定義
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 30 行（実装 20 行 + テスト 10 行）

- [x] **3-2**: [Go] Wikinoのメール送信を事前レンダリング方式に変更

  - ユースケースでテンプレートをレンダリング
  - ジョブにはレンダリング済み文字列を渡す
  - 専用エンキューメソッドを汎用Insertに変更
  - **想定ファイル数**: 約 4 ファイル（実装 3 + テスト 1）
  - **想定行数**: 約 150 行（実装 100 行 + テスト 50 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **ジョブのWeb UI**: River UIの導入は別途検討
- **分散ワーカー**: 複数サーバーでのワーカー実行は現時点では不要
- **優先度キュー**: 現時点ではデフォルトキューのみで十分

## 参考資料

- [River - Fast and reliable background jobs in Go](https://riverqueue.com/)
- [pgx - PostgreSQL Driver and Toolkit for Go](https://github.com/jackc/pgx)
