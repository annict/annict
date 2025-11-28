# Annict 開発ガイド (Go 版)

このファイルは、Go 版 Annict の開発に関するガイダンスを提供します。

> **Note**: プロジェクト全体の概要、モノレポ構造、共通インフラ（PostgreSQL、Redis、imgproxy）については、[/CLAUDE.md](../CLAUDE.md) を参照してください。

## Go 版の開発方針

Go 版は既存の Rails 実装を段階的に再実装しており、以下の方針で開発を進めています：

- **Go 初心者にもわかりやすい実装**: 過度な抽象化を避け、標準ライブラリを優先的に使用
- **YAGNI 原則**: 必要になったときに必要な機能だけを実装

### リバースプロキシによる段階的移行

Go 版と Rails 版を同一ドメイン（`annict.com`）で運用するため、**リバースプロキシミドルウェア**を使用しています：

- **動作**: Go 版で未実装の機能は自動的に Rails 版にプロキシされます
- **メリット**:
  - SEO 対策（`annict.com`と`go.annict.com`の重複を回避）
  - ユーザーは常に`annict.com`でアクセス可能
  - 機能ごとに段階的に移行可能
- **Go 版で処理するパス**: `/static/*`, `/health`, `/manifest.json`, `/sign_in`, `/password/*` など
- **Rails 版にプロキシするパス**: 上記以外のすべてのパス（`/`, `/works`, `/works/popular`, `/@username` など）
- **実装**: `internal/middleware/reverse_proxy.go`（ホワイトリスト方式）

### Rails 版のソースコードを参照する

Go 版の実装時には、Rails 版のコードを参考にして既存の仕様を理解できます。Rails 版のソースコードは `/workspace/rails/` 配下に格納されています。

```
/workspace/rails/
├── app/controllers/     # コントローラー
├── app/models/          # モデル
├── app/views/           # ビューテンプレート
├── config/routes.rb     # ルーティング定義
└── db/structure.sql     # DBスキーマ
```

例：Rails 版のコードを確認する場合

```sh
# Rails版のコントローラーを確認
cat /workspace/rails/app/controllers/works_controller.rb
# Rails版のモデルを確認
cat /workspace/rails/app/models/work.rb
```

## 技術スタック

- Go 1.25.1
  - chi/v5: HTTP ルーターとミドルウェア
  - lib/pq: PostgreSQL ドライバー
  - sqlc: SQL クエリからタイプセーフな Go コードを生成
  - templ: 型安全な HTML テンプレートエンジン
  - resend-go/v2: メール送信ライブラリ（Resend API）
- PostgreSQL 17.3
- Cloudflare Turnstile: Bot 対策サービス（ログイン・サインアップフォームなど）
- pnpm
  - @tailwindcss/cli: Tailwind CSS v4 CLI ツール
  - tailwindcss: Tailwind CSS v4（CSS フレームワーク）
  - esbuild: JavaScript バンドラー
  - basecoat-css: UI コンポーネントライブラリ

## プロジェクト構造

関心の分離を意識した**3層アーキテクチャ**を採用しています。

```
┌─────────────────────────────────────────────────────────┐
│ Presentation層                                          │
│ - Handler, ViewModel, Template, Middleware            │
└─────────────────────────────────────────────────────────┘
         ↓ 依存
┌─────────────────────────────────────────────────────────┐
│ Application層                                           │
│ - UseCase（ビジネスフロー、トランザクション管理）           │
└─────────────────────────────────────────────────────────┘
         ↓ 依存
┌─────────────────────────────────────────────────────────┐
│ Domain/Infrastructure層（統合）                          │
│ - Query (sqlc), Repository, Model                     │
└─────────────────────────────────────────────────────────┘
```

### 主要なパッケージ

- **cmd/server/main.go**: エントリポイント
- **internal/handler**: HTTPリクエストハンドラー（Presentation層）
- **internal/middleware**: HTTPミドルウェア（Presentation層）
- **internal/templates**: templテンプレート（Presentation層）
- **internal/viewmodel**: プレゼンテーション層のデータ変換（Presentation層）
- **internal/usecase**: ビジネスロジック層（Application層）
- **internal/query**: sqlc生成コード（Domain/Infrastructure層）
- **internal/repository**: Repository層（Domain/Infrastructure層）
- **internal/model**: ドメインモデル（Domain/Infrastructure層）
- **internal/config**: 設定管理
- **internal/i18n**: 国際化
- **internal/image**: 画像URL生成
- **internal/session**: セッション管理

### 重要な設計原則

- **依存の方向**: Presentation層 → Application層 → Domain/Infrastructure層
- **Queryへの依存はRepositoryのみ**: Handler/UseCaseがQueryに直接依存することは禁止
- **ModelとRepositoryは1:1の関係**: 各ドメインエンティティに対応するRepositoryを作成
- **Domain/Infrastructure層の統合**: データベース変更はほぼ起こらないため、シンプルさを優先

📖 **詳細なアーキテクチャについては [@go/docs/architecture-guide.md](docs/architecture-guide.md) を参照してください。**

## 開発環境のセットアップ

> **Note**: 開発環境の基本的なセットアップ手順は [/CLAUDE.md](../CLAUDE.md#開発環境のセットアップ) を参照してください。

- Dev Container を使って開発します
- Claude Code はコンテナ内で実行されているため、ホスト側のコマンドの実行は不要です
- 共通インフラ（PostgreSQL、Redis、imgproxy）は `/workspace/docker-compose.yml` で管理されており、ホスト側で起動済みのはずです
- 画像ストレージは Cloudflare R2 を使用（開発環境・本番環境ともに）

### 環境変数の設定

**環境変数の命名規則**:

- Annict で定義する環境変数は、外部ライブラリなどが指定してくるものを除き、**必ずプレフィックス `ANNICT_` を付ける**
- 例:
  - ✅ `ANNICT_PORT`, `ANNICT_DOMAIN`, `ANNICT_RAILS_APP_URL`
  - ❌ `PORT`, `DOMAIN`, `RAILS_APP_URL`
- 外部ライブラリが要求する環境変数はそのまま使用（例: `DATABASE_URL`, `REDIS_URL`）

環境変数の設定には 2 つの方法があります：

#### 方法 1: .env ファイルを使用する

`.env.example` をコピーして `.env` ファイルを作成し、実際の値を設定します：

```sh
# .env ファイルを作成
cp .env.example .env
# エディタで .env を開き、実際の値を設定（API キーなど）
```

`.env` ファイルは `.gitignore` に含まれているため、秘密情報を安全に管理できます。

#### 方法 2: 環境変数を直接設定する

本番環境や CI/CD 環境では、ホスティング環境が提供する環境変数設定機能を使用します：

- **本番環境**: Dokku の環境変数設定（`dokku config:set`）
- **開発環境**: シークレット管理ツールを使用してメモリ上に環境変数を設定

**環境変数の優先順位**（上から順に優先）:

1. **事前に設定された環境変数** - シークレット管理ツールやホスティング環境が提供（メモリ上に存在）
2. **`.env`** - 環境変数設定ファイル（gitignore 対象）

シークレット管理ツール（`op run`）やホスティング環境（GitHub Actions、Dokku）が環境変数をセットするため、Goプロセス起動時には既に環境変数がセット済みです。

### 開発ツールの依存関係管理

Go プロジェクトで使用する CLI ツール（`golangci-lint`, `templ`, `goimports` など）は、`tools.go` ファイルで依存関係を管理します。

#### tools.go ファイルの目的

`tools.go` ファイルは、開発ツールの依存関係を `go.mod` に記録するためのファイルです。以下のメリットがあります：

- **バージョン管理**: `go.mod` でツールのバージョンを固定できる
- **チーム全体で統一**: 全開発者が同じバージョンのツールを使用できる
- **CI との整合性**: ローカル環境と CI 環境で同じバージョンを使用できる
- **自動インストール**: `go mod download` でツールの依存関係も自動的にダウンロードされる

#### 新しい CLI ツールを追加する手順

**1. tools.go にツールを追加**

```go
//go:build tools

package tools

// このファイルは開発ツールの依存関係を go.mod に記録するためのものです。
// ビルドタグ "tools" により、本番ビルドには含まれません。
import (
	_ "github.com/a-h/templ/cmd/templ"
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "github.com/新しいツール/cmd/ツール名"  // 新しいツールを追加
)
```

**2. 依存関係を go.mod に追加**

```sh
# 特定のバージョンを指定する場合
go get github.com/新しいツール/cmd/ツール名@vX.Y.Z

# 最新版を使用する場合
go mod tidy
```

**3. ツールをインストール**

```sh
go install github.com/新しいツール/cmd/ツール名
```

**4. Makefile にターゲットを追加（オプション）**

よく使用するツールは Makefile にターゲットを追加すると便利です：

```makefile
.PHONY: 新しいコマンド
新しいコマンド: ## 新しいツールを実行
	@which ツール名 > /dev/null || (echo "Installing ツール名 from go.mod..." && go install github.com/新しいツール/cmd/ツール名)
	ツール名 [オプション]
```

#### ベストプラクティス

- **ビルドタグの使用**: `//go:build tools` タグにより、本番ビルドにツールが含まれないようにする
- **blank import**: `_ "package"` 形式で import し、実際には使用しない
- **バージョン固定**: 本番環境と同じバージョンを使用するため、バージョンを明示的に指定する
- **Makefile での自動インストール**: ツールがない場合は自動的に `go install` でインストールする仕組みを追加

#### 例: golangci-lint の追加

golangci-lint v2.6.2 を追加する例：

```sh
# 1. tools.go に追加（既に追加済み）
# _ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"

# 2. 特定のバージョンを取得
go get github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2

# 3. インストール
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint

# 4. バージョン確認
golangci-lint version
# => golangci-lint has version 2.6.2 built with go1.25.1 ...
```

### ホスト側で実行するコマンド (Claude Code による実行は不要)

```sh
# コンテナ起動
docker compose up

# コンテナ内に入る
docker compose exec app zsh

# イメージを再構築
docker compose build app --no-cache

# app サービスのログを確認
docker compose logs -f app
```

### コンテナ内で実行するコマンド (Claude Code が実行できるコマンド)

```sh
# 依存関係のインストール
go mod download
pnpm install

# 開発サーバー起動 (ホットリロード)
# 開発サーバーは基本的にホスト側で起動されています
air

# 開発サーバー起動 (ホットリロード無し)
make run

# バイナリビルド (`bin/server` に出力)
make build

# テスト実行（テスト用DBのセットアップも自動実行されます）
make test

# 特定のパッケージのテストを実行（1Password CLI経由で環境変数を自動設定）
make test-pkg PKG=internal/handler/password_reset

# 特定のテストを実行（1Password CLI経由で環境変数を自動設定）
make test-run PKG=internal/handler/password_reset RUN=TestCreate_TurnstileVerification

# 詳細ログ付きで全テストを実行（デバッグ用）
make test-verbose

# テスト用DBのセットアップのみ実行
make db-setup-test

# コードフォーマット
make fmt

# リント
make lint

# ビルド成果物のクリーン
make clean

# sqlc コード生成
sqlc generate

# templ コード生成（go.modのバージョンを自動使用）
# 注: templ-generateは自動的にgoimportsも実行してimport文を整理します
make templ-generate

# goimportsでimport文を整理（go.modのバージョンを自動使用）
# 注: 通常はtempl-generateやfmtが自動的に実行するため、手動実行は不要です
make goimports
# import文のチェック（差分があればエラー）
make goimports-check

# PostgreSQL (開発環境用) に接続する
# Claude CodeはDockerコンテナ内で動いているため、postgresqlホスト名でアクセス可能
psql -h postgresql -p 5432 -U postgres -d annict_development

# PostgreSQL (テスト環境用) に接続する
psql -h postgresql -p 5432 -U postgres -d annict_test

# フロントエンドアセットのビルド
pnpm build       # CSS と JS をビルド（本番用、minify 有効）
pnpm build:css   # CSS のみビルド
pnpm build:js    # JS のみビルド

# フロントエンドアセットの監視（開発時）
pnpm watch       # CSS と JS を監視して自動再ビルド
pnpm watch:css   # CSS のみ監視
pnpm watch:js    # JS のみ監視
```

### コミット前に実行するコマンド

**重要**: コードをコミットする前に、以下のコマンドを実行して CI が通ることを確認してください：

```sh
# 0. .templファイルを編集した場合は、Goコードを再生成（自動的にgoimportsも実行されます）
make templ-generate

# 1. go mod tidyを実行して依存関係を整理
go mod tidy

# 2. コードフォーマット（go fmt, goimportsを含む）
make fmt

# 3. 静的解析（golangci-lint）を実行
# golangci-lintは複数のリンターを統合して実行します:
#   - gofmt: フォーマットチェック
#   - govet: Go標準静的解析
#   - staticcheck: 高度な静的解析
#   - gosec: セキュリティチェック
#   - depguard: アーキテクチャルール（依存関係チェック）
#   - errcheck: エラーチェック漏れ
#   - ineffassign: 無駄な代入
#   - unused: 未使用コード
make lint

# 4. ビルドを実行（全パッケージの型チェック）
# 関数シグネチャの変更やインポートエラーを早期に検出
go build ./...

# 5. テストを実行
APP_ENV=test make test

# すべてを一度に実行するワンライナー:
make templ-generate && go mod tidy && make fmt && make lint && go build ./... && APP_ENV=test make test
```

#### golangci-lint の使い方

golangci-lint は複数の静的解析ツールを統合して実行するリンターです。

**基本的な使い方**:

```sh
# 全パッケージに対してリントを実行
make lint

# 特定のパッケージに対してリントを実行
golangci-lint run --config=.golangci.yml ./internal/handler/...

# 特定のリンターのみ実行
golangci-lint run --config=.golangci.yml --disable-all --enable=gosec ./...
```

**設定ファイル**: `.golangci.yml`

golangci-lint の詳細な設定は `.golangci.yml` で管理されています。主な設定内容：

- **有効化されているリンター**: gofmt, govet, staticcheck, gosec, depguard, errcheck, ineffassign, unused
- **アーキテクチャルール（depguard）**: 3層アーキテクチャの依存関係を強制
  - Presentation層（Handler, Middleware, ViewModel, Templates）はQueryに直接依存できない
  - Application層（UseCase）はQueryに直接依存できない
  - Domain/Infrastructure層（Query, Repository, Model）は上位層に依存できない

**CI での実行**:

GitHub Actions で自動的に golangci-lint が実行されます。ローカルで `make lint` を実行してエラーがなければ、CI も通過します。

### データベースマイグレーション

このプロジェクトでは [dbmate](https://github.com/amacneil/dbmate) を使用してデータベースマイグレーションを管理しています。

#### 基本的な使い方

```sh
# 新しいマイグレーションファイルを作成
make db-new name=create_users

# マイグレーションを実行（開発環境）
make db-migrate

# テスト用データベースのセットアップ
make db-setup-test

# 最後のマイグレーションをロールバック
make db-rollback

# DBスキーマをダンプ
make db-dump

# sqlcでGoコードを生成
make sqlc-generate
```

#### 注意事項

- マイグレーションファイルは `db/migrations/` ディレクトリに作成されます
- スキーマは `db/schema.sql` に出力されます
- **テスト実行時**: `make test` を実行すると、自動的に `make db-setup-test` が実行されるため、手動でセットアップを実行する必要はありません
- **テスト用 DB のセットアップ方法**:
  - テスト用 DB (`annict_test`) を完全にリセット（`DROP SCHEMA public CASCADE`）
  - `db/schema.sql` を適用して最新のスキーマを再作成
  - これにより、常にクリーンな状態でテストが実行されます
- PostgreSQL 17.6 の `\restrict` コマンド対策として、自動的にクリーンアップ処理が実行されます
  - 参照: https://github.com/amacneil/dbmate/issues/678

## Pull Request のガイドライン

Pull Request のガイドラインは [/CLAUDE.md](../CLAUDE.md#pull-requestのガイドライン) を参照してください。

**要約**:

- 実装コード: 300 行以下を目安
- テストコード: 制限なし（必要な分だけ書く）
- 実装とテストは同じ PR に含める
- 「行数を守ること」よりも「きちんと実装すること」を優先

## コーディング規約

### Go コード

- **インデント**: タブを使用（Go 標準）
- **フォーマット**: `gofmt` を使用して自動フォーマット
- **コメント**: 自動生成されたコード以外は日本語でコメントを記述する
  - 関数やメソッドの説明は日本語で記述
  - インラインコメントも日本語で統一
  - sqlc 等のツールが生成したコードのコメントは変更しない

#### コメントのガイドライン

**良いコメント**：

- コードの**意図や理由**を説明する（「なぜこうしたか」）
- 将来の開発者が理解できる、文脈に依存しない内容
- 複雑なロジックや、一見不自然に見える実装の背景を説明する

```go
// 良い例: 意図を説明
// ユーザーが削除済みでも、過去の記録との整合性を保つためにIDは保持する
if user.DeletedAt != nil {
    return user.ID, nil
}

// 良い例: 制約や前提条件を説明
// NOTE: PostgreSQL 17.6のバグ回避のため、このクエリでは明示的なCAST(id AS bigint)が必要
// https://github.com/postgres/postgres/issues/12345
query := "SELECT CAST(id AS bigint) FROM users"
```

**避けるべきコメント**：

- ❌ **実装の変遷を説明するコメント**（「以前は〜だった」「〜は削除した」など）
- ❌ **過去との比較**（「別途インストール不要になった」「〜を統合したため不要」など）
- ❌ **自明なことの説明**（コードを読めばわかること）
- ❌ **やり取りの文脈に依存するコメント**（PR レビューのコメントは PR に書く）

```go
// 悪い例: 実装の変遷を説明（git履歴で確認できる）
// 以前はここでgo installしていたが、makeに統合したため削除した

// 悪い例: 自明なことを説明
// ユーザーIDを取得
userID := user.ID

// 良い例: 複雑なロジックの意図を説明
// ユーザーIDを取得（削除済みユーザーは0を返す）
userID := user.ID
if user.DeletedAt != nil {
    userID = 0
}
```

**コメントが不要な場合**：

- コードが十分に明確で、読めばわかる場合
- 適切な変数名・関数名で意図が伝わる場合

```go
// 不要: コードを読めば明らか
// make templ-generateを実行
make templ-generate

// 必要なし: 関数名で意図が明確
func generateTemplFiles() error {
    return exec.Command("make", "templ-generate").Run()
}
```

**原則**：

- **コメントはコードの「なぜ」を説明し、「何を」はコードに語らせる**
- git の履歴に残る情報（過去の実装、変更の経緯）はコメントに書かない
- レビューコメントや議論の文脈に依存する内容は書かない

#### ログ出力

ログ出力は Go 1.21 で標準ライブラリに追加された `log/slog` パッケージを使用します。

**基本ルール**:

- ✅ **`log/slog` を使用する**: 構造化ログで一貫性のあるログ出力を実現
- ❌ **`log` パッケージは使用禁止**: `log.Printf`, `log.Println`, `log.Fatalf` は使用しない

**使用する関数**:

| 関数 | 用途 |
|------|------|
| `slog.Info(msg, key, value, ...)` | 通常の情報ログ（コンテキストなし） |
| `slog.Warn(msg, key, value, ...)` | 警告ログ（コンテキストなし） |
| `slog.Error(msg, key, value, ...)` | エラーログ（コンテキストなし） |
| `slog.InfoContext(ctx, msg, key, value, ...)` | 通常の情報ログ（コンテキストあり） |
| `slog.WarnContext(ctx, msg, key, value, ...)` | 警告ログ（コンテキストあり） |
| `slog.ErrorContext(ctx, msg, key, value, ...)` | エラーログ（コンテキストあり） |

**コンテキストの使い分け**:

- **コンテキストが利用可能な場合**（Handler, UseCase など）: `slog.InfoContext(ctx, ...)` を使用
- **コンテキストが利用不可能な場合**（`main.go` など）: `slog.Info(...)` を使用

**ログレベルの選択基準**:

| レベル | 用途 | 例 |
|--------|------|-----|
| `slog.Debug` | デバッグ情報（通常は出力しない） | 変数の値、処理の詳細 |
| `slog.Info` | 通常の情報（サーバー起動、処理完了など） | サーバー起動、DB接続成功 |
| `slog.Warn` | 警告（処理は継続するが注意が必要） | 非推奨機能の使用、リトライ |
| `slog.Error` | エラー（処理が失敗した場合） | DB接続失敗、API呼び出し失敗 |

**良い例**:

```go
// コンテキストありの場合（Handler, UseCase）
slog.InfoContext(ctx, "ユーザーがログインしました", "user_id", userID)
slog.ErrorContext(ctx, "パスワードリセットメールの送信に失敗", "error", err, "email", email)

// コンテキストなしの場合（main.go）
slog.Info("サーバーを起動します", "port", cfg.Port)
slog.Error("データベース接続に失敗", "error", err)
```

**悪い例（使用禁止）**:

```go
// ❌ log パッケージは使用禁止
log.Printf("ユーザーがログインしました: %d", userID)
log.Println("サーバーを起動します")
log.Fatalf("データベース接続に失敗: %v", err)
```

**致命的エラーの処理**:

`log.Fatalf` の代わりに `slog.Error` + `os.Exit(1)` を使用します。

```go
// ❌ 悪い例
log.Fatalf("データベース接続に失敗: %v", err)

// ✅ 良い例
slog.Error("データベース接続に失敗", "error", err)
os.Exit(1)
```

### HTTP ハンドラー

HTTP リクエストを処理するハンドラーは、統一された規則に従って実装します。

#### 基本方針

- **リソースごとにディレクトリを切る**: すべてのエンドポイントは `internal/handler/{resource}/` 配下にディレクトリを作成
- **1 エンドポイント = 1 ハンドラーファイル**: 各エンドポイントは個別のファイルに実装
- **統一された命名規則**: ファイル名とメソッド名に一貫性を持たせる
- **例外なくディレクトリ化**: 単独のエンドポイントでも必ずディレクトリを作成（例: `health/`, `home/`）

#### 標準ファイル名（8 種類のみ）

リソースディレクトリ内には、以下の標準的なファイル名**のみ**を使用します：

- `handler.go` - Handler 構造体と依存性の定義
- `index.go` - 一覧ページ表示 (GET /resources)
- `show.go` - 個別リソースの表示（フォームなし） (GET /resources/:id)
- `new.go` - 新規作成フォーム表示（作成前のフォーム） (GET /resources/new)
- `create.go` - 作成処理（新規リソースの永続化） (POST /resources)
- `edit.go` - 編集フォーム表示（更新前のフォーム） (GET /resources/:id/edit)
- `update.go` - 更新処理（既存リソースの永続化） (PATCH /resources/:id)
- `delete.go` - 削除処理 (DELETE /resources/:id)

**重要な区別**:

- **`new.go`**: 新規作成フォームを表示し、その後 `create.go` で永続化する（例: ユーザー登録フォーム → ユーザー作成）
- **`edit.go`**: 編集フォームを表示し、その後 `update.go` で永続化する（例: プロフィール編集フォーム → プロフィール更新）
- **`show.go`**: フォームを含まない、既存リソースの詳細表示のみ（例: ユーザープロフィール表示）

#### メソッド名

ファイル名とメソッド名は完全に対応します：

- `Index` - 一覧ページ表示（ファイル名: `index.go`）
- `Show` - 個別リソースの表示（フォームなし）（ファイル名: `show.go`）
- `New` - 新規作成フォーム表示（作成前のフォーム）（ファイル名: `new.go`）
- `Create` - 作成処理（新規リソースの永続化）（ファイル名: `create.go`）
- `Edit` - 編集フォーム表示（更新前のフォーム）（ファイル名: `edit.go`）
- `Update` - 更新処理（既存リソースの永続化）（ファイル名: `update.go`）
- `Delete` - 削除処理（ファイル名: `delete.go`）

**重要**: 複雑なメソッド名（`ShowResetForm`, `ProcessReset` など）は使用しません。代わりに新しいリソースディレクトリを作成し、標準的なメソッド名を使用します。

#### 詳細ドキュメント

ハンドラーの詳しい実装方法、ディレクトリ構造、依存性注入のガイドライン、実装例については以下を参照してください：

📖 **[@go/docs/handler-guide.md](docs/handler-guide.md)** - HTTP ハンドラーガイドライン

### templ テンプレート

Go 版では、型安全なテンプレートエンジン [templ](https://templ.guide/) を使用して HTML を生成します。

#### 基本情報

- **ファイル配置**: `internal/templates/` ディレクトリ
  - `layouts/`: レイアウトテンプレート
  - `components/`: 再利用可能なコンポーネント
  - `pages/`: ページテンプレート
  - `emails/`: メールテンプレート
- **命名規則**: 小文字のスネークケース（例: `sign_in.templ`, `popular.templ`）
- **生成ファイル**: `*_templ.go` ファイルは自動生成されるため、**手動編集禁止**
- **コード生成**: `templ generate` で Go コードを生成

#### 主な特徴

- **型安全性**: コンパイル時に型チェックとエラー検出
- **コンポーネント化**: `@componentName()` で他のコンポーネントを呼び出し
- **国際化対応**: `templates.T(ctx, "message_id")` で翻訳を取得

#### 詳細ドキュメント

テンプレートの詳しい書き方、レイアウトの継承、コンポーネントの再利用、テストの書き方などは以下のドキュメントを参照してください：

📖 **[@go/docs/templ-guide.md](docs/templ-guide.md)** - templ テンプレートガイド

### リクエストバリデーション（Request Validation）

フォームからの入力値の検証は、**Request DTO（Data Transfer Object）パターン**を使用します。

#### 基本情報

- **責務**: リクエストデータの構造定義と**形式バリデーションのみ**を行う
- **ファイル配置**: ハンドラーと同じディレクトリ
- **命名規則**: `{Action}Request`
- **バリデーション範囲**: 必須チェック、フォーマット検証、文字数制限など（DB アクセス不要）

#### 詳細ドキュメント

バリデーションの詳しい実装方法、複雑なバリデーション例、テストの書き方については以下を参照してください：

📖 **[@go/docs/validation-guide.md](docs/validation-guide.md)** - リクエストバリデーションガイド

### HTTP メソッドとルーティング

HTML フォームと Web API（JSON）で同じルーティングを使用するため、**Method Override パターン**を採用します：

#### 基本方針

- **Web API（JSON）**: 標準的な HTTP メソッド（GET/POST/PATCH/DELETE）を使用
- **HTML フォーム**: POST に`_method`パラメータを追加して PATCH/DELETE を実現
- **ルーティング**: HTML と API で同じ URL・HTTP メソッドを共有
- **更新処理**: PATCH を使用（部分更新を表現、Rails との整合性も保つ）

#### 実装方法

**ミドルウェア**: `internal/middleware/method_override.go`

```go
r.Use(authMiddleware.MethodOverride) // POSTリクエストの_methodパラメータを読み取り、HTTPメソッドを上書き
```

**HTML フォームの例**:

```html
<!-- PATCH /password として処理される -->
<form method="POST" action="/password">
  <input type="hidden" name="_method" value="PATCH" />
  <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
  <!-- フォームフィールド -->
</form>
```

**Web API（JSON）の例**:

```javascript
// そのままPATCH /passwordとして処理される
fetch("/password", {
  method: "PATCH",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ password: "newpass" }),
});
```

**ルーティング**:

```go
// HTMLフォームもJSON APIも同じルーティング
r.Patch("/password", h.UpdatePassword)
r.Patch("/users/{id}", h.PartialUpdateUser)
r.Delete("/posts/{id}", h.DeletePost)
```

#### メリット

- HTML と API で同じ URL・HTTP メソッドを使用できる
- RESTful 設計を維持できる
- 将来の Web API 実装が容易（ルーティング変更不要）
- Rails 方式と同じため、既存の Rails アプリからの移行もスムーズ

### 国際化（I18n）

すべてのユーザー向けメッセージは**必ず国際化対応**します。

#### 基本情報

- **対応言語**: 日本語（デフォルト）と英語
- **翻訳ファイル**: `internal/i18n/locales/ja.toml` と `internal/i18n/locales/en.toml`
- **使用方法**:
  - テンプレート: `templates.T(ctx, "message_id")`
  - Go コード: `i18n.T(ctx, "message_id")`
- **対象**: ページタイトル、ラベル、エラーメッセージなど
- **対象外**: ログメッセージ、panic メッセージ、内部エラー（開発者向けは日本語で OK）

#### 詳細ドキュメント

翻訳の追加手順、命名規則、プレースホルダーの使い方、テストの書き方については以下を参照してください：

📖 **[@go/docs/i18n-guide.md](docs/i18n-guide.md)** - 国際化（I18n）ガイド

### ビューモデル・ユースケース

Go 版 Annict では、関心の分離を意識したアーキテクチャを採用しています。

#### ビューモデル（View Model）

プレゼンテーション層でのデータ変換を担当します。

- **配置**: `internal/viewmodel`
- **責務**: リポジトリ層のデータをテンプレート表示用に変換
- **命名**: `NewWorkFromXXX`, `NewWorksFromXXX`

#### ユースケース（Use Case）

ビジネスロジックとトランザクション管理を担当します。

- **配置**: `internal/usecase` （フラット構造）
- **責務**: トランザクション管理、複数リポジトリを跨ぐ処理
- **命名**: ファイル名 `{action}_{entity}.go`、構造体名 `{Action}{Entity}Usecase`

#### 詳細ドキュメント

ビューモデルとユースケースの詳しい実装方法、命名規則、テストの書き方については以下を参照してください：

📖 **[@go/docs/architecture-guide.md](docs/architecture-guide.md)** - アーキテクチャガイド

## セキュリティガイドライン

Web アプリケーションのセキュリティは**最優先事項**です。

### 基本対策

- **CSRF 対策**: すべてのフォーム送信に CSRF トークンを含める
- **XSS 対策**: templ の自動エスケープを活用、ユーザー入力を信頼しない
- **SQL インジェクション対策**: sqlc のプリペアドステートメントを使用
- **パスワード管理**: bcrypt でハッシュ化、平文はログに出力しない
- **認証・認可**: ログインチェックとリソース所有者チェックを実施
- **エラーメッセージ**: 詳細な情報を漏らさない、ログに記録
- **Bot 対策**: Cloudflare Turnstile でログイン・サインアップフォームを保護

### セキュリティチェックリスト

新機能を実装する際は、以下を必ず確認してください：

- [ ] フォーム送信に CSRF トークンを含めているか
- [ ] ユーザー入力をバリデーションしているか
- [ ] SQL クエリはプリペアドステートメントを使用しているか
- [ ] パスワードは bcrypt でハッシュ化されているか
- [ ] 認証・認可チェックを行っているか
- [ ] エラーメッセージは適切か

### 詳細ドキュメント

セキュリティ対策の詳しい実装方法、具体例、トラブルシューティングについては以下を参照してください：

📖 **[@go/docs/security-guide.md](docs/security-guide.md)** - セキュリティガイドライン

## テスト戦略

### 基本方針

- **実データベースを使用**: 基本的にデータベースをモックせず、実際の PostgreSQL データベースを使用してテストを実行
- **トランザクションでの分離**: 各テストはトランザクション内で実行し、テスト終了時に自動ロールバックすることでデータをクリーンアップ
- **テストヘルパーの活用**: `internal/testutil` パッケージのヘルパー関数とビルダーパターンを使用してテストデータを作成
- **自動スキーマセットアップ**: `make test` を実行すると、テスト用データベースが自動的にリセットされ、`db/schema.sql` が適用されます（常にクリーンな状態でテストが実行される）

### テストの構造

- **テストファイル**: `*_test.go` という接尾辞を付けて同じディレクトリに配置
- **テスト関数**: `Test` で始まる名前（例: `TestPopularWorks`）
- **ベンチマーク関数**: `Benchmark` で始まる名前（例: `BenchmarkPopularWorks`）

### テストのベストプラクティス

- **実データベースを使用**: モックではなく実際の PostgreSQL データベースでテスト
- **トランザクション分離**: `testutil.SetupTestDB(t)` でテスト用 DB とトランザクションをセットアップ
- **テーブル駆動テスト**: 複数のテストケースを効率的に実行
- **並行テスト**: `t.Parallel()` で並行実行可能なテストを高速化（トランザクション分離により安全）
- **テストヘルパー**: 共通のセットアップコードをヘルパー関数に抽出
- **エラーケースを必ずテスト**: 正常系だけでなく異常系も網羅

### 実データベーステストの例

```go
func TestPopularWorks(t *testing.T) {
    // テストDBとトランザクションをセットアップ
    db, tx := testutil.SetupTestDB(t)

    // テストデータを作成（ビルダーパターン）
    workID := testutil.NewWorkBuilder(t, tx).
        WithTitle("テストアニメ").
        WithSeason(2024, testutil.SeasonSpring).
        Build()

    // sqlcリポジトリを作成（トランザクションを使用）
    queries := repository.New(db).WithTx(tx)

    // ハンドラーを作成してテスト実行
    handler := &Handler{
        queries: queries,
        cfg:     cfg,
        templates: templates,
    }

    // HTTPリクエストとレスポンスのテスト
    req := httptest.NewRequest("GET", "/works/popular", nil)
    rr := httptest.NewRecorder()
    handler.PopularWorks(rr, req)

    // アサーション
    if rr.Code != http.StatusOK {
        t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusOK)
    }

    // テスト終了時にトランザクションは自動的にロールバックされる
}
```

### テストヘルパーの使用

`internal/testutil` パッケージには以下のヘルパーが用意されています：

- **`SetupTestDB(t)`**: テスト用データベース接続とトランザクションのセットアップ
- **`NewWorkBuilder(t, tx)`**: 作品データのビルダー
- **`NewEpisodeBuilder(t, tx, workID)`**: エピソードデータのビルダー
- **`NewUserBuilder(t, tx)`**: ユーザーデータのビルダー
- **`NewWorkImageBuilder(t, tx, workID)`**: 作品画像データのビルダー
- **`CreateTestWork(t, tx, title)`**: 簡易的な作品作成ヘルパー

### 個別テスト実行タスク

開発効率を向上させるため、特定のパッケージやテストのみを実行する Makefile タスクを用意しています。これらのタスクは自動的に以下を実行します：

- **1Password CLI のラッパー**: `op run --env-file=".env"` を経由して環境変数を自動設定
- **DB セットアップ**: テスト実行前に自動的に `db-setup-test` を実行

#### 利用可能なタスク

**`make test-pkg PKG=<パッケージ>`**

特定のパッケージのテストを実行します。

```sh
# パスワードリセットハンドラーのテストを実行
make test-pkg PKG=internal/handler/password_reset

# Turnstileパッケージのテストを実行
make test-pkg PKG=internal/turnstile

# 認証パッケージのテストを実行
make test-pkg PKG=internal/auth
```

**`make test-run PKG=<パッケージ> RUN=<テスト名>`**

特定のパッケージの特定のテストを実行します（パターンマッチング可能）。

```sh
# Turnstile検証のテストのみ実行
make test-run PKG=internal/handler/password_reset RUN=TestCreate_TurnstileVerification

# 特定のテストケースのみ実行
make test-run PKG=internal/handler/password_reset RUN=TestCreate_TurnstileVerification_Success

# パターンマッチでRateLimiting関連のテストを実行
make test-run PKG=internal/handler/password_reset RUN=TestCreate_RateLimiting
```

**`make test-verbose`**

詳細ログ付きで全テストを実行します（デバッグ用）。キャッシュを無効化して実行されます。

```sh
# キャッシュを無効化して全テストを詳細ログ付きで実行
make test-verbose
```

#### 使用例

開発中に特定の機能のテストのみを素早く実行する場合：

```sh
# 1. パッケージ全体のテストを実行
make test-pkg PKG=internal/handler/password_reset

# 2. 特定のテストが失敗した場合、そのテストのみを実行
make test-run PKG=internal/handler/password_reset RUN=TestCreate_TurnstileVerification_Failed

# 3. 修正後、再度パッケージ全体のテストを実行して確認
make test-pkg PKG=internal/handler/password_reset
```

**メリット**:

- **短いコマンド**: 長い `APP_ENV=test op run --env-file=".env" -- go test -v -race ./...` を打つ必要がない
- **環境変数の自動設定**: 1Password CLI を経由して自動的に環境変数が設定される
- **DB セットアップの自動化**: テスト実行前に自動的にテスト用 DB がセットアップされる
- **エラーハンドリング**: PKG や RUN 変数が指定されていない場合、使用例を表示

### テンプレートレンダリングのテスト

templ テンプレートのレンダリングも含めたテストを実装します。templ は型安全なため、コンパイル時に多くのエラーを検出できますが、実際の HTML 出力もテストする必要があります。

#### 基本方針

- **ハンドラーテスト**: HTTP リクエスト・レスポンスを含めた統合テスト
- **コンポーネント単体テスト**: `Render()` メソッドを使って直接テスト
- **テーブル駆動テスト**: 複数のロケールやケースをまとめてテスト
- **HTML 出力の検証**: 特定の要素やテキストが正しくレンダリングされているか確認

#### 詳細ドキュメント

templ テンプレートのテストの詳しい書き方、具体的なコード例、ベストプラクティスについては以下のドキュメントを参照してください：

📖 **[@go/docs/templ-guide.md](docs/templ-guide.md)** - templ テンプレートガイド（テストセクション）

---

## 関連ドキュメント

- **プロジェクト全体のガイド**: [/workspace/CLAUDE.md](../CLAUDE.md) - モノレポ構造、共通インフラ、Rails から Go への移行について
- **Rails 版のガイド**: [/workspace/rails/CLAUDE.md](../rails/CLAUDE.md) - Rails 版の技術スタック、開発環境、コーディング規約
