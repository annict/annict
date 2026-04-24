---
paths:
  - "go/**"
---

# Wikino 開発ガイド (Go 版)

このファイルは、Go 版 Wikino の開発に関するガイダンスを提供します。

> **Note**: プロジェクト全体の概要、モノレポ構造、共通インフラ（PostgreSQL）については、[/CLAUDE.md](../CLAUDE.md) を参照してください。

## Go 版の開発方針

Go 版は既存の Rails 実装を段階的に再実装しており、以下の方針で開発を進めています：

- **Go 初心者にもわかりやすい実装**: 過度な抽象化を避け、標準ライブラリを優先的に使用

### リバースプロキシによる段階的移行

Go 版で未実装の機能は自動的に Rails 版にプロキシされます（`internal/middleware/reverse_proxy.go`、ホワイトリスト方式）。

### Rails 版のソースコードを参照する

Go 版の実装時には、Rails 版のコード（`/workspace/rails/`）を参考にして既存の仕様を理解できます。

## 技術スタック

- Go 1.26.2
  - chi/v5: HTTP ルーターとミドルウェア
  - lib/pq: PostgreSQL ドライバー
  - sqlc: SQL クエリからタイプセーフな Go コードを生成
  - templ: 型安全な HTML テンプレートエンジン
  - resend-go/v2: メール送信ライブラリ（Resend API）
  - river: バックグラウンドジョブキュー（PostgreSQL ベース）
- PostgreSQL 18.1
- htmx 4: ハイパーメディアフレームワーク（HTML フラグメント返却によるサーバードリブン UI）
  - htmx の実装時は `/htmx4` スキルを使用すること
- Cloudflare Turnstile: Bot 対策サービス
- pnpm
  - @tailwindcss/cli + tailwindcss: Tailwind CSS v4
  - esbuild: JavaScript バンドラー
  - basecoat-css: UI コンポーネントライブラリ

## プロジェクト構造

関心の分離を意識した**3 層アーキテクチャ**を採用しています。

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

### 主要なパッケージ

| パッケージ            | 層                    | 責務                       |
| --------------------- | --------------------- | -------------------------- |
| `cmd/server/main.go`  | -                     | エントリポイント           |
| `internal/handler`    | Presentation          | HTTP リクエストハンドラー  |
| `internal/middleware` | Presentation          | HTTP ミドルウェア          |
| `internal/templates`  | Presentation          | templ テンプレート         |
| `internal/viewmodel`  | Presentation          | 表示用データ変換           |
| `internal/worker`     | Presentation          | バックグラウンドジョブ受信 |
| `internal/usecase`    | Application           | オーケストレーション       |
| `internal/validator`  | Application           | 入力バリデーション         |
| `internal/query`      | Domain/Infrastructure | sqlc 生成コード            |
| `internal/repository` | Domain/Infrastructure | Repository 層              |
| `internal/model`      | Domain/Infrastructure | ドメインモデル             |
| `internal/policy`     | Domain/Infrastructure | 認可ロジック               |
| `internal/dispatcher` | Domain/Infrastructure | ジョブキュー投入           |
| `internal/email`      | Presentation          | メール送信・レンダリング   |
| `internal/config`     | -                     | 設定管理                   |
| `internal/i18n`       | -                     | 国際化                     |
| `internal/session`    | -                     | セッション管理             |

### 重要な設計原則

- **依存の方向**: Presentation 層 → Application 層 → Domain/Infrastructure 層
- **Handler から Repository への直接依存は禁止**: Handler のすべてのデータアクセスは UseCase を経由する
- **UseCase はオーケストレーター**: 書き込み UseCase は認可・バリデーション・ビジネスロジック・永続化を統括する。読み取り UseCase はデータ取得を担当する
- **Handler / Worker は薄い Adapter**: HTTP/ジョブの入出力変換のみ。認可・バリデーションは UseCase を経由する
- **Validator は Application 層**: すべてのバリデーションは `internal/validator/` に配置し、UseCase から呼び出される
- **認可チェックは UseCase で実行**: UseCase が Policy を呼び出して認可チェックを行う。Handler から Policy への直接依存は禁止
- **Query への依存は Repository のみ**: Handler/UseCase/Worker が Query に直接依存することは禁止
- **Model と Repository は 1:1 の関係**: 各ドメインエンティティに対応する Repository を作成
- **ドメイン ID 型の使用**: モデルの ID フィールドには `string` ではなく専用のドメイン ID 型を使用する

📖 **詳細は [docs/architecture-guide.md](docs/architecture-guide.md) を参照**

## 開発コマンド

> **Note**: Claude Code はコンテナ内で実行されています。共通インフラ（PostgreSQL）はホスト側で起動済みです。

```sh
# 依存関係のインストール
go mod download && pnpm install

# 開発サーバー起動
air                    # ホットリロード
make run               # ホットリロードなし

# ビルド
make build             # bin/server に出力

# テスト実行
make test                                         # 全テスト
make test-pkg PKG=internal/handler/password_reset  # パッケージ指定
make test-run PKG=internal/handler/password_reset RUN=TestCreate  # テスト指定
make test-verbose                                  # 詳細ログ付き

# コードフォーマット・リント
make fmt               # go fmt + goimports
make lint              # golangci-lint
make templ-generate    # templ → Go コード生成（goimports も自動実行）

# データベース
make db-migrate        # マイグレーション実行
make db-new name=xxx   # 新しいマイグレーション作成
make sqlc-generate     # sqlc コード生成

# PostgreSQL に接続
psql -h postgresql -p 5432 -U postgres -d wikino_development  # 開発
psql -h postgresql -p 5432 -U postgres -d wikino_test          # テスト

# フロントエンドアセット
pnpm build             # CSS + JS ビルド（本番用）
pnpm watch             # CSS + JS 監視（開発用）
```

### コミット前に実行するコマンド

```sh
# すべてを一度に実行するワンライナー:
make templ-generate && go mod tidy && make fmt && make lint && go build ./... && make -C /workspace fmt && make test
```

個別に実行する場合：

1. `make templ-generate` - .templ ファイルを編集した場合
2. `go mod tidy` - 依存関係を整理
3. `make fmt` - コードフォーマット
4. `make lint` - 静的解析（golangci-lint）
5. `go build ./...` - 全パッケージの型チェック
6. `make -C /workspace fmt` - JS/TS/Markdown ファイルを編集した場合
7. `make test` - テスト実行

## ガイドライン一覧

各トピックの詳細は以下のドキュメントを参照してください：

| ドキュメント                                             | 内容                                           |
| -------------------------------------------------------- | ---------------------------------------------- |
| [docs/architecture-guide.md](docs/architecture-guide.md) | 3 層アーキテクチャ、Repository、Worker         |
| [docs/usecase-guide.md](docs/usecase-guide.md)           | UseCase の設計と実装パターン                   |
| [docs/handler-guide.md](docs/handler-guide.md)           | HTTP ハンドラー、ルーティング、Method Override |
| [docs/coding-guide.md](docs/coding-guide.md)             | コーディング規約、コメント、ログ出力           |
| [docs/templ-guide.md](docs/templ-guide.md)               | templ テンプレート、ViewModel との関係         |
| [docs/validation-guide.md](docs/validation-guide.md)     | バリデーション                                 |
| [docs/i18n-guide.md](docs/i18n-guide.md)                 | 国際化（I18n）                                 |
| [docs/security-guide.md](docs/security-guide.md)         | セキュリティ                                   |
| [docs/testing-guide.md](docs/testing-guide.md)           | テスト戦略                                     |
| [docs/development-guide.md](docs/development-guide.md)   | 開発環境、DB マイグレーション、golangci-lint   |

## Pull Request のガイドライン

[/CLAUDE.md](../CLAUDE.md#pull-requestのガイドライン) を参照。要約：実装コード 300 行以下目安、テストコード制限なし、実装とテストは同じ PR に含める。
