# Annict 開発ガイドライン

## プロジェクト概要

Annict はアニメ視聴記録サービスです。
ユーザーは自分が見たアニメに対して「見てる」や「見たい」といったステータスを設定したり、見たアニメの感想を書いてあとから振り返ることができます。

## モノレポ構造

このリポジトリは、Go 版と Rails 版の 2 つのサブプロジェクトをモノレポとして管理しています:

```
/workspace/
├── go/                  # Go版の実装（段階的に機能を移行中）
├── rails/               # Rails版の実装（既存の本番システム）
├── caddy/               # リバースプロキシ設定（Caddy）
├── imgproxy/            # imgproxy 設定
├── docs/                # Annict固有のドキュメント（仕様書、作業計画書など）
├── .claude/             # AIガイドライン・スキル（apm install で自動配置）
├── .github/             # 共通のCI/CD設定
├── apm.yml              # APM（Agent Package Manager）の依存関係定義
├── apm.lock.yaml        # APM のロックファイル
├── apm_modules/         # APM の作業ディレクトリ
├── Dockerfile.dev       # 統合開発コンテナの Dockerfile
├── docker-compose.yml   # Docker Compose 設定
├── Makefile             # 開発タスクのエントリポイント
├── Procfile.dev         # hivemind による開発サーバープロセス定義
├── mise.toml            # 開発ツールバージョン管理
└── CLAUDE.md            # このファイル（プロジェクト全体のガイド）
```

## Rails から Go への移行について

現在、既存の Rails 実装の Annict を Go で段階的に再実装するプロジェクトが進行中です。

### 移行の基本方針

- **既存 DB をそのまま使用**: Rails 側で管理されている PostgreSQL データベースを共有
- **段階的移行**: Rails と Go が同一の DB とセッションストアを共有し、段階的に機能を移行
- **データマイグレーション不要**: DB スキーマは既存のものを使用し、データ移行は行わない
- **共通インフラの継続利用**: 画像配信 (Cloudflare R2 + imgproxy) などの共通インフラは Go 版移行後も継続して使用
- **ルーティング振り分け**: Caddy と Go 側のリバースプロキシミドルウェアで、Go 版で未実装の機能は自動的に Rails 版にプロキシされる

### Rails 版の主要技術スタック

Rails 版で利用している主な gem と役割 (Go 版への移行対象を把握するための参考情報):

- **認証**: Devise
- **認可**: Pundit
- **OAuth**: Doorkeeper
- **GraphQL API**: graphql-ruby (`app/graphql/`)
- **バックグラウンドジョブ**: Delayed Job
- **画像アップロード**: Shrine + Cloudflare R2
- **ビューコンポーネント**: ViewComponent
- **テンプレート**: Slim / ERB
- **E2E テスト**: RSpec + Capybara + Playwright (`spec/system/`)

### Rails 側のソースコード

Rails 版のソースコードは `/workspace/rails/` 配下に格納されています:

```
/workspace/rails/
├── app/controllers/     # コントローラー
├── app/models/          # モデル
├── app/views/           # ビューテンプレート
├── config/routes.rb     # ルーティング定義
└── db/structure.sql     # DB スキーマ
```

Go 版を実装する際は、Rails 版のコードを参考にすることで既存の仕様を理解できます。

## 共通インフラ

### データベース (PostgreSQL)

- **バージョン**: PostgreSQL 17.x
- **共有方針**: Rails 版と Go 版で同一のデータベースを共有
- **開発環境**: Docker Compose で管理 (ホスト側ポート: 4001 / コンテナ内: 5432)
- **データベース名**:
  - 開発: `annict_development`
  - テスト: `annict_test`

### セッションストア (PostgreSQL)

- **ストレージ**: PostgreSQL の `sessions` テーブルを使用
- **Rails 版**: ActiveRecord SessionStore (セッション有効期限 30 日、`updated_at` を各リクエストで自動更新)
- **Go 版**: 同じ `sessions` テーブルを共有し、認証ミドルウェアで `updated_at` を更新 (Rails 版と完全に互換)
- **セッションクリーンアップ**: 毎日 19:00 に `rake session:sweep` タスクが実行され、30 日以上前のセッションを自動削除
- **共有方針**: Rails 版と Go 版で同一のセッションストアを共有することで段階的移行を実現

### 画像配信 (Cloudflare R2 + imgproxy)

作品画像の配信は Rails 版と Go 版で共通利用します。

- **オブジェクトストレージ**: Cloudflare R2 (S3 互換)
  - バケット: 開発 `annict-development` / 本番 `annict-production`
  - 画像パスは `shrine/` プレフィックスで保存 (Rails Shrine の仕様)
- **imgproxy**: 画像リサイズ・最適化プロキシ
  - ポート: 18080 (開発環境)
  - S3 プロトコル経由でストレージにアクセス (`s3://annict-{environment}/shrine/{path}`)
  - 署名付き URL を生成してセキュアに配信 (KEY/SALT は環境変数で管理)
  - Docker Compose で管理
- **共有方針**: Go 版が本流になってもこの構成を継続利用する

## 開発環境のセットアップ

### 前提条件

- Docker および Docker Compose がインストール済み
- Dev Container を使用した開発環境

### セットアップ手順

1. リポジトリをクローン
2. VS Code や Claude Code でリポジトリを開くと、Dev Container が自動的に起動
3. ホスト側で `docker compose up` を実行し、共通インフラ (PostgreSQL、imgproxy など) を起動

Go / Rails 固有のセットアップ手順 (依存関係のインストール、マイグレーション、テスト用 DB の初期化など) は `.claude/rules/go-development.md` と `.claude/rules/rails-common.md` を参照してください。

### 開発サーバーの起動

プロジェクトルートで以下のコマンドを実行すると、Go 版・Rails 版の全サービスを一括で起動できます:

```sh
make dev
```

このコマンドは [hivemind](https://github.com/DarthSim/hivemind) を使用して `Procfile.dev` に定義された以下のプロセスを並行起動します:

| プロセス       | 内容                                        |
| -------------- | ------------------------------------------- |
| `go-assets`    | Go 版フロントエンドアセットの監視・再ビルド |
| `go-server`    | Go 版サーバー (air によるホットリロード)    |
| `rails-css`    | Rails 版 CSS の監視・再ビルド               |
| `rails-js`     | Rails 版 JavaScript の監視・再ビルド        |
| `rails-server` | Rails 版サーバー                            |
| `rails-worker` | Rails 版バックグラウンドワーカー            |

### 環境変数の命名規則

Annict では、自前で定義する環境変数のプレフィックスに **`ANNICT_`** を使用します。共通ガイドの `.claude/rules/go-development.md` では `WIKINO_` が例示されていますが、Annict では `ANNICT_` に読み替えてください。

- 代表例: `ANNICT_PORT`, `ANNICT_DOMAIN`, `ANNICT_RAILS_APP_URL`
- 例外: `APP_ENV` はプレフィックスなしで使用
- 外部ライブラリが要求する環境変数 (例: `DATABASE_URL`) はそのまま使用する

## ドキュメント

ドキュメントは `docs/` 配下で管理しており、ユーザーが直接体験する機能の仕様は `docs/specs/` に、ユーザーが直接体験しないシステム内部の仕組みは `docs/system/` に配置しています。配置先の判断基準やディレクトリ構成は [@docs/README.md](/workspace/docs/README.md) を参照してください。

- [@docs/README.md](/workspace/docs/README.md) - ドキュメント全体のガイド
- [@docs/specs/](/workspace/docs/specs/) - サービス仕様書 (ユーザーが直接体験する機能)
- [@docs/system/](/workspace/docs/system/) - システム仕様書 (ユーザーが直接体験しないシステム内部の仕組み)

## 参照するガイドライン

Claude Code は `.claude/rules/` 配下のガイドラインを自動で読み込むため、通常は特に意識せず書いて OK。ガイドラインの実体は `korylus-guidelines` から `apm install` で配置されています。

- **Korylus 共通**: `.claude/rules/common.md` / `.claude/rules/apm.md`
- **Go 版**: `.claude/rules/go-*.md`
- **Rails 版**: `.claude/rules/rails-*.md`

APM 管理下のファイルは `apm install` で上書きされます。編集したい場合は `korylus-guidelines` 側を修正してください。

## 開発ワークフロー

### フィーチャーフラグによる開発

Korylus 共通の方針は [@.claude/rules/common.md](/workspace/.claude/rules/common.md) の「フィーチャーフラグによる開発」セクションを参照してください。

Annict における具体的な仕組み (DB スキーマ、リバースプロキシの判定ロジックなど) は仕様書を参照:

- [@docs/specs/feature-flag/overview.md](/workspace/docs/specs/feature-flag/overview.md) — フィーチャーフラグ 仕様書

## CI/CD

このモノレポの CI/CD 設定は `.github/workflows/` ディレクトリに配置されています:

- `go-ci.yml`: Go 版の CI (lint、test、build)
- `rails-ci.yml`: Rails 版の CI (zeitwerk、sorbet、standard、erb_lint、eslint、rspec)
- `fmt-ci.yml`: フォーマットチェック (Oxfmt)

各 CI は対応するファイルが変更されたときに実行されます (パスフィルタリング)。

## トラブルシューティング

### データベース接続エラー

- PostgreSQL コンテナが起動しているか確認: `docker compose ps`
- ポートが正しいか確認: ホスト側からは 4001、コンテナ内からは 5432

### 画像が表示されない

- imgproxy コンテナが起動しているか確認: `docker compose ps`
- 環境変数 (R2 アクセスキー、imgproxy KEY/SALT) が正しく設定されているか確認
- Cloudflare R2 バケットへのアクセス権限が正しく設定されているか確認
