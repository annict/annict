# 環境変数の統合（`ANNICT_ENV` から `APP_ENV` へ）設計書

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

Go プロジェクト内に `ANNICT_ENV` と `APP_ENV` という2つの環境変数が存在し、どちらも環境（development, test, production）を表しています。この2つの環境変数を `APP_ENV` に統一することで、コードの一貫性を向上させ、Annict 以外のプロジェクトでも同じパターンを適用できるようにします。

**目的**:

- 環境変数の命名を統一し、コードの一貫性を向上させる
- Annict 固有のプレフィックスを削除し、汎用的な命名にする
- 1Password CLI（`op run`）との統合をより明確にする

**背景**:

- 現在、`ANNICT_ENV` は Go アプリケーション内部で環境識別のために使用されている（単一の `.env` ファイルを読み込む）
- 一方、`APP_ENV` は Makefile や `.air.toml` で 1Password CLI（`op run`）に渡され、`.env` ファイル内で環境別のシークレットを参照するために使用されている（`env_$APP_ENV`）
- 両者は同じ概念（開発環境、テスト環境、本番環境）を表しているが、異なる名前を使用しており、混乱を招く可能性がある
- `APP_ENV` は汎用的な命名であり、Annict 以外のプロジェクトでも使いやすい

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

- `internal/config/config.go` は `APP_ENV` 環境変数を読み取り、環境識別に使用する
- `APP_ENV` の値：
  - `dev`: 開発環境
  - `test`: テスト環境
  - `prod`: 本番環境
- **重要**: Go側で`.env`ファイルを読み込む処理は不要
  - ローカル開発/テスト: `op run --env-file=".env"` が`.env`を読み込んで環境変数をセット済み
  - CI環境: GitHub Actionsが`env:`で環境変数をセット済み
  - 本番環境: Dokkuが環境変数をセット済み
  - すべての環境でGoプロセス起動時には既に環境変数がセット済みのため、`godotenv.Load()`は不要
- `APP_ENV` が未設定の場合は、デフォルトで `dev` として扱う
- 既存のテストが正常に動作する
- Makefile、`.air.toml`、`.env.example` は既に `APP_ENV` を使用しているため、変更不要
- ドキュメント（CLAUDE.md）内の `ANNICT_ENV` の記述を `APP_ENV` に更新する

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

- **下位互換性**: `ANNICT_ENV` は完全に削除し、`APP_ENV` のみをサポートする（後方互換性は不要）
- **保守性**: コードの一貫性を向上させ、将来の開発者が理解しやすくする
- **汎用性**: Annict 固有のプレフィックスを削除し、他のプロジェクトでも使いやすい命名にする

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

### 現状の環境変数の使用状況

#### `ANNICT_ENV` の使用箇所

1. **`go/internal/config/config.go`**: 環境識別のために使用（`.env` ファイルは環境に関係なく単一のファイルのみ読み込む）
   ```go
   // ANNICT_ENVの値を取得（デフォルト: development）
   env := os.Getenv("ANNICT_ENV")
   if env == "" {
       env = "development"
   }

   // 単一の .env ファイルを読み込む（環境別ファイルではない）
   _ = godotenv.Load(".env")
   ```

2. **`go/internal/config/config_test.go`**: テストで環境を指定するために使用
   ```go
   os.Setenv("ANNICT_ENV", "development")
   // または
   os.Setenv("ANNICT_ENV", "test")
   ```

3. **`go/CLAUDE.md`**: ドキュメント内でテスト実行時に使用
   ```sh
   ANNICT_ENV=test make test
   ANNICT_ENV=test go test -v ./internal/handler -run TestPopularWorksWithRealDB
   ```

#### `APP_ENV` の使用箇所

1. **`go/.air.toml`**: 開発サーバー（air）でのビルド時に `op run` コマンドに渡される
   ```toml
   cmd = "make templ-generate && make goimports && APP_ENV=dev op run --env-file=.env -- go build -o ./tmp/main ./cmd/server"
   full_bin = "APP_ENV=dev op run --env-file=.env -- ./tmp/main"
   ```

2. **`go/Makefile`**: `op run` コマンドに渡される
   ```makefile
   run: ## サーバーを起動
       APP_ENV=dev op run --env-file=".env" -- go run cmd/server/main.go

   test: ## テストを実行
       APP_ENV=test op run --env-file=".env" -- go test -v -race ./...

   db-migrate: ## データベースマイグレーションを実行
       APP_ENV=dev op run --env-file=".env" -- dbmate up
   ```

3. **`go/.env.example`**: 1Password の環境変数参照で使用
   ```sh
   ANNICT_DOMAIN=op://VAULT_EXAMPLE/env_$APP_ENV/Go/ANNICT_DOMAIN
   ANNICT_IMGPROXY_KEY=op://VAULT_EXAMPLE/env_$APP_ENV/Go/ANNICT_IMGPROXY_KEY
   ```

### 統合方針

- **目標**: `ANNICT_ENV` を完全に削除し、`APP_ENV` のみを使用する
- **環境変数の管理方法**:
  - 環境別に複数の`.env`ファイルを用意するのではなく、単一の`.env`ファイルのみを使用
  - `.env`ファイル内で1Password CLIの環境変数参照（`env_$APP_ENV`）を使用し、環境別のシークレットを取得
  - **Go側で`.env`ファイルを読み込む処理は不要**：
    - ローカル開発/テスト: `op run --env-file=".env"` が`.env`を処理してGoプロセスを起動
    - CI環境: GitHub Actionsの`env:`設定で環境変数をセット
    - 本番環境: Dokkuで環境変数をセット
  - すべての環境でGoプロセス起動時には既に環境変数がセット済み
  - この構成により、Go側のコードが簡素化され、環境ごとの分岐処理も不要になる
- **利点**:
  - 環境変数が1つになり、混乱を避けられる
  - `APP_ENV` は汎用的な命名であり、他のプロジェクトでも使いやすい
  - Makefile や `.air.toml` が既に `APP_ENV` を使用しているため、整合性が取れる
  - 1Password CLI との統合がより明確になる

### 変更内容

#### 1. `internal/config/config.go`

`ANNICT_ENV` を `APP_ENV` に変更し、`.env`ファイルの読み込み処理を削除：

```go
// APP_ENVの値を取得（デフォルト: dev）
// dev: 開発環境、test: テスト環境、prod: 本番環境
//
// 注: すべての環境でGoプロセス起動時には既に環境変数がセット済みのため、
// .envファイルの読み込み処理は不要です：
// - ローカル開発/テスト: op run --env-file=".env" が処理済み
// - CI環境: GitHub Actionsが設定済み
// - 本番環境: Dokkuが設定済み
env := os.Getenv("APP_ENV")
if env == "" {
    env = "dev"
}
```

また、`godotenv`パッケージのimportも削除：

```go
import (
    "fmt"
    "os"
    "os/exec"
    "strconv"
    "strings"
    "time"
    // "github.com/joho/godotenv" は削除
)
```

また、環境判定メソッドも修正：

```go
// IsDev は開発環境かどうかを返します
func (c *Config) IsDev() bool {
    return c.Env == "dev"
}

// IsProduction は本番環境かどうかを返します
func (c *Config) IsProduction() bool {
    return c.Env == "prod"
}
```

#### 2. `go.mod`

`godotenv`パッケージへの依存を削除：

```bash
# go.modから自動的に削除される
go mod tidy
```

#### 3. `internal/config/config_test.go`

すべての `ANNICT_ENV` を `APP_ENV` に変更し、環境値も`dev`/`test`/`prod`に統一：

```go
os.Setenv("APP_ENV", "dev")
// または
os.Setenv("APP_ENV", "test")

// クリーンアップ
defer os.Unsetenv("APP_ENV")
```

また、`.env`ファイルの読み込みに関連するテストを削除：
- `TestLoad_PresetEnvironmentVariables` テストは削除（`.env`読み込み処理がなくなるため不要）

#### 4. `.github/workflows/go-ci.yml`

GitHub Actions の環境変数を `ANNICT_ENV` → `APP_ENV` に更新：

```yaml
env:
  APP_ENV: test  # ANNICT_ENV から変更
  DATABASE_URL: postgres://postgres@localhost:5432/annict_test?sslmode=disable
  # ... その他の環境変数
```

#### 5. `go/CLAUDE.md`

ドキュメント内の `ANNICT_ENV` の記述を `APP_ENV` に更新：

```sh
APP_ENV=test make test
APP_ENV=test go test -v ./internal/handler -run TestPopularWorksWithRealDB
```

### テスト戦略

- 既存のテストが正常に動作することを確認
- 特に `internal/config/config_test.go` のテストが正常に動作することを確認
- 環境変数が未設定の場合にデフォルトで `dev` が使用されることを確認
- `.env`ファイル読み込みに関連するテストの削除：
  - `TestLoad_PresetEnvironmentVariables` テストを削除（`.env`読み込み処理がなくなるため不要）

### 影響範囲

- **影響あり**:
  - `internal/config/config.go`
    - `ANNICT_ENV` → `APP_ENV` に変更
    - デフォルト値を `"development"` → `"dev"` に変更
    - `.env` ファイルの読み込みロジックを**完全に削除**（`godotenv.Load()`の呼び出しを削除）
    - `godotenv`パッケージのimportを削除
    - 関連するコメントを簡素化
    - `IsDev()` メソッドを `"development"` → `"dev"` に変更
    - `IsProduction()` メソッドを `"production"` → `"prod"` に変更
  - `go.mod`
    - `go mod tidy`により`godotenv`パッケージへの依存が自動削除される
  - `internal/config/config_test.go`
    - すべての `ANNICT_ENV` → `APP_ENV` に変更
    - テストで使用する環境値を `"development"` → `"dev"` に変更
    - `TestLoad_PresetEnvironmentVariables` テストを削除
  - `.github/workflows/go-ci.yml`
    - `ANNICT_ENV: test` → `APP_ENV: test` に変更
  - `go/CLAUDE.md`
    - ドキュメント内の `ANNICT_ENV=test` → `APP_ENV=test` に更新

- **影響なし（既に `APP_ENV` を使用）**:
  - `go/Makefile`
  - `go/.air.toml`
  - `go/.env.example`

- **確認が必要**:
  - `"development"`, `"production"` という文字列が他のコードで使用されていないか確認
  - `godotenv`パッケージが他の場所で使用されていないか確認

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

### フェーズ 1: 環境変数の統合

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: `ANNICT_ENV` を `APP_ENV` に統合

  - `internal/config/config.go` の変更:
    - `ANNICT_ENV` → `APP_ENV` に変更
    - デフォルト値を `"development"` → `"dev"` に変更
    - `.env` ファイルの読み込みロジックを**完全に削除**（`godotenv.Load()`を削除）
    - `godotenv`パッケージのimportを削除
    - 関連するコメントを簡素化
    - `IsDev()` メソッドを `"development"` → `"dev"` に変更
    - `IsProduction()` メソッドを `"production"` → `"prod"` に変更
  - `go.mod` の変更:
    - `go mod tidy`により`godotenv`パッケージへの依存が自動削除される
  - `internal/config/config_test.go` の変更:
    - すべての `ANNICT_ENV` → `APP_ENV` に変更
    - テストで使用する環境値を `"development"` → `"dev"` に変更
    - `TestLoad_PresetEnvironmentVariables` テストを削除
  - `.github/workflows/go-ci.yml` の変更:
    - `ANNICT_ENV: test` → `APP_ENV: test` に変更
  - `go/CLAUDE.md` のドキュメント更新:
    - すべての `ANNICT_ENV=test` → `APP_ENV=test` に更新
  - テストを実行し、すべてのテストが正常に動作することを確認
  - **想定ファイル数**: 約 5 ファイル（実装 3 + CI設定 1 + ドキュメント 1）
  - **想定行数**: 約 100 行（実装 60 行 + CI設定 10 行 + ドキュメント 30 行）
    - 削除: 約 20 行（`.env`読み込みロジック、`godotenv` import、テスト削除）
    - 追加: 約 10 行（コメント修正、環境判定メソッド修正）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **後方互換性の維持**: `ANNICT_ENV` のサポートは完全に削除します。既存の環境変数を使用している場合は、`APP_ENV` に変更する必要があります。

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Twelve-Factor App - Config](https://12factor.net/config) - 環境変数の設計原則
- [1Password CLI - Environment Variables](https://developer.1password.com/docs/cli/reference/commands/run/) - `op run` コマンドの環境変数の使い方

---

## 調査結果

### `ANNICT_ENV` の使用状況と役割

1. **`go/internal/config/config.go`** (go/internal/config/config.go:69-78):
   ```go
   // ANNICT_ENVの値を取得（デフォルト: development）
   env := os.Getenv("ANNICT_ENV")
   if env == "" {
       env = "development"
   }

   // .envファイルを読み込む（不要な処理）
   _ = godotenv.Load(".env")
   ```

   **現在の問題点**:
   - `"development"` という値を使用しているが、Makefileでは `dev` を使用
   - `.env`ファイルの読み込み処理が含まれているが、すべての環境でGoプロセス起動時には既に環境変数がセット済みのため不要
     - ローカル開発/テスト: `op run --env-file=".env"` が処理済み
     - CI環境: GitHub Actionsが設定済み
     - 本番環境: Dokkuが設定済み
   - 環境判定メソッド（`IsDev()`, `IsProduction()`）も `"development"`, `"production"` を使用

2. **`go/internal/config/config_test.go`** - 4箇所で使用:
   - テストセットアップ時に `os.Setenv("ANNICT_ENV", "development")` または `os.Setenv("ANNICT_ENV", "test")`
   - テスト終了時に `os.Unsetenv("ANNICT_ENV")`

3. **`go/CLAUDE.md`** - 4箇所で使用:
   - テスト実行例: `ANNICT_ENV=test make test`
   - テスト実行例: `ANNICT_ENV=test go test -v ./internal/handler -run TestPopularWorksWithRealDB`

4. **`.github/workflows/go-ci.yml`** - 1箇所で使用:
   - テストジョブの環境変数: `ANNICT_ENV: test`

### `APP_ENV` の使用状況と役割

1. **`go/.air.toml`** - 2箇所で使用:
   - `cmd = "make templ-generate && make goimports && APP_ENV=dev op run --env-file=.env -- go build -o ./tmp/main ./cmd/server"`
   - `full_bin = "APP_ENV=dev op run --env-file=.env -- ./tmp/main"`

   **役割**: 開発サーバー（air）起動時に `APP_ENV=dev` を1Password CLIに渡す

2. **`go/Makefile`** - 6箇所で使用:
   - `run`: `APP_ENV=dev op run --env-file=".env" -- go run cmd/server/main.go`
   - `test`: `APP_ENV=test op run --env-file=".env" -- go test -v -race ./...`
   - `db-migrate`: `APP_ENV=dev op run --env-file=".env" -- dbmate up`
   - `db-rollback`: `APP_ENV=dev op run --env-file=".env" -- dbmate down`
   - `db-dump`: `APP_ENV=dev op run --env-file=".env" -- dbmate dump`
   - `seed`: `APP_ENV=dev op run --env-file=".env" -- go run cmd/seed/main.go`

   **役割**: 各コマンド実行時に `APP_ENV=dev` または `APP_ENV=test` を1Password CLIに渡す

3. **`go/.env.example`** - 複数箇所で使用:
   - `ANNICT_DOMAIN=op://VAULT_EXAMPLE/env_$APP_ENV/Go/ANNICT_DOMAIN`
   - `ANNICT_IMGPROXY_KEY=op://VAULT_EXAMPLE/env_$APP_ENV/Go/ANNICT_IMGPROXY_KEY`
   - など、1Password の環境変数参照で使用

   **役割**: `.env`ファイル内で`env_$APP_ENV`として1Passwordの環境別シークレットを参照

### 統合の必要性

- `ANNICT_ENV` と `APP_ENV` は同じ概念（環境識別）を表しているが、異なる値（`development` vs `dev`, `production` vs `prod`）を使用
- Makefileと`.air.toml`は既に`APP_ENV`を使用しているため、Go コード側も統一すべき
- **重要**: `.env`ファイルの読み込み処理（`godotenv.Load()`）は完全に不要
  - すべての環境でGoプロセス起動時には既に環境変数がセット済み：
    - ローカル開発/テスト: `op run --env-file=".env"` が処理
    - CI環境: GitHub Actionsが設定
    - 本番環境: Dokkuが設定
  - Go側のコードを簡素化でき、環境ごとの分岐処理も不要になる
