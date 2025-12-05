# Annict 開発ガイド (Rails版)

このファイルは、Rails版Annictの開発に関するガイダンスを提供します。

> **Note**: プロジェクト全体の概要、モノレポ構造、共通インフラ（PostgreSQL、Redis、imgproxy）については、[/CLAUDE.md](../CLAUDE.md) を参照してください。

## Rails版の開発方針

Rails版は既存の本番システムであり、以下の方針で開発・保守を進めています：

- **安定性優先**: 既存機能の動作を維持しながら慎重に改善
- **段階的なGo版移行**: 機能ごとに段階的にGo版へ移行
- **共通インフラの継続利用**: PostgreSQL、Redis、imgproxyなどはGo版と共有

### Go版への移行について

現在、Rails版の機能を段階的にGo版へ移行中です。Go版の実装時には、Rails版のコードを参考にして既存の仕様を理解できます。

## 技術スタック

### バックエンド

- Ruby 3.3.10
- Rails 7.1.6
- PostgreSQL 17.3
- Redis
- **認証**: Devise
- **認可**: Pundit
- **OAuth**: Doorkeeper
- **GraphQL**: graphql-ruby (~> 2.0.32)
- **バックグラウンドジョブ**: Delayed Job
- **画像アップロード**: Shrine + Cloudflare R2
- **画像処理**: imgproxy
- **型チェック**: Sorbet
- **リント**: Standard (RuboCop)
- **テスト**: RSpec

### フロントエンド

- **JavaScriptフレームワーク**: Hotwire (Stimulus, Turbo)
- **CSSフレームワーク**: Bootstrap 5
- **CSSプリプロセッサ**: Sass
- **バンドラー**: esbuild
- **パッケージマネージャー**: Yarn
- **リント**: ESLint (TypeScript ESLint)
- **フォーマッター**: Prettier
- **テンプレート**: Slim, ERB
- **コンポーネント**: ViewComponent

### その他

- **国際化**: Rails I18n (日本語・英語)
- **メール送信**: Resend
- **エラー追跡**: Sentry
- **ページネーション**: Kaminari
- **マークダウン**: CommonMarker + GitHub Markup

## プロジェクト構造

Rails標準のMVCアーキテクチャに加え、サービス層やコンポーネントを導入した構造：

```
/workspace/rails/
├── app/
│   ├── controllers/      # コントローラー（HTTPリクエスト処理）
│   ├── models/           # モデル（ActiveRecord）
│   ├── views/            # ビューテンプレート（Slim/ERB）
│   ├── components/       # ViewComponent（再利用可能なUIコンポーネント）
│   ├── services/         # サービスオブジェクト（ビジネスロジック）
│   ├── queries/          # クエリオブジェクト（複雑なDB検索）
│   ├── policies/         # 認可ロジック（Pundit）
│   ├── graphql/          # GraphQL API定義
│   ├── forms/            # フォームオブジェクト
│   ├── decorators/       # デコレーター（ActiveDecorator）
│   ├── helpers/          # ビューヘルパー
│   ├── javascript/       # JavaScriptファイル
│   ├── assets/           # CSS、画像などのアセット
│   ├── jobs/             # バックグラウンドジョブ
│   ├── mailers/          # メーラー
│   └── uploaders/        # Shrineアップローダー
├── config/
│   ├── routes.rb         # ルーティング定義
│   ├── application.rb    # アプリケーション設定
│   ├── database.yml      # データベース設定
│   ├── initializers/     # 初期化処理
│   └── locales/          # 国際化ファイル
├── db/
│   ├── migrate/          # マイグレーションファイル
│   └── structure.sql     # DBスキーマ（PostgreSQL形式）
├── spec/                 # RSpecテスト
├── sorbet/               # Sorbet型定義
├── public/               # 静的ファイル
├── bin/                  # 実行可能スクリプト
├── Gemfile               # Ruby依存関係
├── package.json          # Node.js依存関係
└── Rakefile              # Rakeタスク
```

## 開発環境のセットアップ

> **Note**: 開発環境の基本的なセットアップ手順は [/CLAUDE.md](../CLAUDE.md#開発環境のセットアップ) を参照してください。

- Dev Containerを使って開発します
- Claude Codeはコンテナ内で実行されているため、ホスト側のコマンドの実行は不要です
- 共通インフラ（PostgreSQL、Redis、imgproxy）は `/docker-compose.yml` で管理されており、ホスト側で起動済みのはずです

### 環境変数の設定

環境変数は`.env.{environment}`ファイルで管理します：

- `.env.development` - 開発環境用
- `.env.test` - テスト環境用
- `.env.production` - 本番環境用（Dokku環境変数で設定）

### ホスト側で実行するコマンド (Claude Codeによる実行は不要)

```sh
# コンテナ起動
cd /workspace/rails
docker compose up -d

# コンテナ内に入る
docker compose exec app bash

# イメージを再構築
docker compose build app --no-cache

# appサービスのログを確認
docker compose logs -f app
```

### コンテナ内で実行するコマンド (Claude Codeが実行できるコマンド)

環境変数の読み込みが必要なコマンドは **Makefile** でラップされています。
`make help` で利用可能なコマンド一覧を確認できます。

```sh
# 依存関係のインストール
bundle install
yarn install

# 開発サーバー起動
# 開発サーバーは基本的にホスト側で起動されています
bin/dev

# Railsサーバーのみ起動
make server

# コンソール起動
make console

# テスト実行
make test
# 特定のテストを実行
make test-file FILE=spec/models/work_spec.rb
# E2Eテストを実行（Playwright）
make test-file FILE=spec/system/

# コードフォーマット
make fmt                           # Ruby（自動修正）
yarn prettier --write "**/*.js"    # JavaScript

# リント
make lint                          # Ruby
yarn eslint "**/*.js"              # JavaScript

# Sorbet型チェック
make sorbet

# Zeitwerk（オートロード）チェック
make zeitwerk

# PostgreSQL（開発環境）に接続
psql -h host.docker.internal -p 15432 -U postgres -d annict_development

# PostgreSQL（テスト環境）に接続
psql -h host.docker.internal -p 15432 -U postgres -d annict_test

# データベースマイグレーション
make db-migrate
make db-rollback    # 最後のマイグレーションをロールバック

# データベースのセットアップ
make db-setup       # DBの作成、スキーマ読み込み、シード実行

# フロントエンドアセットのビルド
yarn build       # JavaScript（本番用、minify有効）
yarn build:css   # CSS（本番用）

# GraphQL APIスキーマのダンプ
make graphql-dump
```

### コミット前に実行するコマンド

**重要**: コードをコミットする前に、以下のコマンドを実行してCIが通ることを確認してください：

```sh
# 1. 型の更新
make sorbet-update

# 2. Zeitwerk（オートロード）チェック
make zeitwerk

# 3. Sorbet型チェック
make sorbet

# 4. Rubyコードのリント・フォーマット
make fmt

# 5. JavaScriptのリント
yarn eslint "**/*.js"

# 6. テストを実行
make test

# すべてを一度に実行するワンライナー:
make sorbet-update && make zeitwerk && make sorbet && make fmt && yarn eslint "**/*.js" && make test
```

## Pull Requestのガイドライン

Pull Requestのガイドラインは [/CLAUDE.md](../CLAUDE.md#pull-requestのガイドライン) を参照してください。

**要約**:

- 実装コード: 300行以下を目安
- テストコード: 制限なし（必要な分だけ書く）
- 実装とテストは同じPRに含める
- 「行数を守ること」よりも「きちんと実装すること」を優先

## コーディング規約

### Rubyコード

- **インデント**: 2スペースを使用（Ruby標準）
- **スタイルガイド**: Standard（RuboCop）に従う
- **自動フォーマット**: `make fmt`を使用
- **コメント**: 日本語で記述（複雑なロジックの説明）
- **型注釈**: Sorbetの型注釈を可能な限り追加

  ```ruby
  # typed: true
  extend T::Sig

  sig { params(user_id: Integer).returns(User) }
  def find_user(user_id)
    User.find(user_id)
  end
  ```

#### コメントのガイドライン

**良いコメント**：

- コードの**意図や理由**を説明する（「なぜこうしたか」）
- 将来の開発者が理解できる、文脈に依存しない内容
- 複雑なロジックや、一見不自然に見える実装の背景を説明する

```ruby
# 良い例: 意図を説明
# ユーザーが削除済みでも、過去の記録との整合性を保つためにIDは保持する
return user.id if user.deleted_at.present?

# 良い例: 制約や前提条件を説明
# NOTE: ActiveRecord 7.0のバグ回避のため、明示的なjoinsが必要
# https://github.com/rails/rails/issues/12345
User.joins(:posts).where(posts: { published: true })
```

**避けるべきコメント**：

- ❌ **実装の変遷を説明するコメント**（「以前は〜だった」「〜は削除した」など）
- ❌ **過去との比較**（「bundle installに統合したため不要」など）
- ❌ **自明なことの説明**（コードを読めばわかること）
- ❌ **やり取りの文脈に依存するコメント**（PR レビューのコメントは PR に書く）

```ruby
# 悪い例: 実装の変遷を説明（git履歴で確認できる）
# 以前はここでGemをインストールしていたが、Gemfileに統合したため削除した

# 悪い例: 自明なことを説明
# ユーザーIDを取得
user_id = user.id

# 良い例: 複雑なロジックの意図を説明
# ユーザーIDを取得（削除済みユーザーは0を返す）
user_id = user.deleted_at.present? ? 0 : user.id
```

**原則**：

- **コメントはコードの「なぜ」を説明し、「何を」はコードに語らせる**
- git の履歴に残る情報（過去の実装、変更の経緯）はコメントに書かない
- レビューコメントや議論の文脈に依存する内容は書かない

詳細については、[/CLAUDE.md](../CLAUDE.md#コメントのガイドライン) を参照してください。

### テンプレート（Slim/ERB）

- **インデント**: 2スペースを使用
- **推奨**: 新規作成はSlimを使用（可読性が高い）
- **ERB**: 既存のERBファイルは無理に変換しない
- **リント**: `bundle exec erblint --lint-all`でチェック

### JavaScript/TypeScript

- **インデント**: 2スペースを使用
- **スタイルガイド**: ESLint (TypeScript ESLint) に従う
- **フォーマッター**: Prettier
- **フレームワーク**: Stimulus Controllerを優先的に使用
- **型定義**: TypeScriptの型定義を可能な限り追加

### アーキテクチャパターン

Rails版Annictは、標準のMVCアーキテクチャに加え、以下のパターンを導入しています：

#### ViewComponent

再利用可能なUIコンポーネントを実装します。

- **配置**: `app/components/`
- **命名**: `{ComponentName}Component`
- **テンプレート**: Slimを使用

#### サービスオブジェクト

複雑なビジネスロジックとトランザクション管理を担当します。

- **配置**: `app/services/`
- **命名**: `{Action}{Entity}Service`
- **メソッド**: `call` メソッドを実装

#### Pundit（認可）

認可ロジックを管理します。

- **配置**: `app/policies/`
- **命名**: `{Model}Policy`
- **コントローラー**: `authorize` メソッドで認可チェック

#### 詳細ドキュメント

各パターンの詳しい実装方法、ベストプラクティス、テストの書き方については以下を参照してください：

📖 **[@rails/docs/architecture-guide.md](docs/architecture-guide.md)** - アーキテクチャガイド

### 国際化（I18n）

すべてのユーザー向けメッセージは**必ず国際化対応**します：

- **対応言語**: 日本語（デフォルト）と英語
- **翻訳ファイル**: `config/locales/ja.yml` と `config/locales/en.yml`
- **ビュー**: `t('.message_key')` または `I18n.t('message_key')` で翻訳を呼び出す
- **モデル**: `human_attribute_name` でカラム名を国際化
- **対象メッセージ**:
  - ✅ ページタイトル、見出し、ラベル、ボタンテキスト
  - ✅ エラーメッセージ、成功メッセージ
  - ✅ ヘルプテキスト、説明文
  - ❌ ログメッセージ（開発者向けのため日本語のままでOK）
  - ❌ 開発者向けエラー（raise、内部エラーなど）

#### バリデーションエラーメッセージの国際化

ActiveRecordのバリデーションエラーメッセージも国際化します：

```yaml
# config/locales/ja.yml
ja:
  activerecord:
    errors:
      models:
        user:
          attributes:
            email:
              blank: "を入力してください"
              invalid: "の形式が正しくありません"
```

## テスト戦略

Rails版Annictは、RSpecを使用した包括的なテストを実施しています。

### 基本方針

- **テストファースト**: 実装前にテストを書くことを推奨
- **実データベースを使用**: 基本的にデータベースをモックせず、実際のPostgreSQLを使用
- **FactoryBot**: テストデータはFactoryBotで作成
- **カバレッジ**: SimpleCovでカバレッジを測定

### テストの種類

- **モデルテスト**: `spec/models/` - バリデーション、メソッドの動作確認
- **コントローラーテスト**: `spec/requests/` - HTTPリクエスト・レスポンス、認証・認可
- **システムテスト**: `spec/system/` - ブラウザを使ったE2Eテスト（Capybara + Playwright）
- **GraphQL APIテスト**: `spec/graphql/` - クエリ・ミューテーションのテスト

### 詳細ドキュメント

各テストタイプの詳しい実装方法、ベストプラクティス、トラブルシューティングについては以下を参照してください：

📖 **[@rails/docs/testing-guide.md](docs/testing-guide.md)** - テスト戦略ガイド

## セキュリティガイドライン

Webアプリケーションのセキュリティは**最優先事項**です。

### 基本対策

- **CSRF 対策**: `protect_from_forgery` がデフォルトで有効、`form_with` ヘルパーを使用
- **XSS 対策**: ERB/Slimの自動エスケープを活用、`raw`/`html_safe` は慎重に使用
- **SQL インジェクション対策**: ActiveRecordのプリペアドステートメント、プレースホルダーを使用
- **認証**: Deviseで管理
- **認可**: Punditで管理
- **Strong Parameters**: すべてのコントローラーで使用

### 詳細ドキュメント

セキュリティ対策の詳しい実装方法、具体例、トラブルシューティングについては以下を参照してください：

📖 **[@rails/docs/security-guide.md](docs/security-guide.md)** - セキュリティガイドライン

## データベース管理

### マイグレーション

```sh
# 新しいマイグレーションを作成
bin/rails generate migration CreateWorks

# マイグレーションを実行
make db-migrate

# マイグレーションをロールバック
make db-rollback

# スキーマをダンプ（structure.sql）
make db-migrate
```

### スキーマ管理

- **スキーマファイル**: `db/structure.sql` （PostgreSQL形式）
- **Go版との共有**: Rails版のマイグレーションがDBスキーマを管理

## GraphQL API

### スキーマ定義

- **配置**: `app/graphql/`
- **スキーマファイル**: `app/graphql/annict_schema.rb`
- **型定義**: `app/graphql/types/`
- **クエリ**: `app/graphql/queries/`
- **ミューテーション**: `app/graphql/mutations/`

### スキーマのダンプ

```sh
make graphql-dump
```

## 関連ドキュメント

- **プロジェクト全体のガイド**: [/CLAUDE.md](../CLAUDE.md) - モノレポ構造、共通インフラ、RailsからGoへの移行について
- **Go版のガイド**: [/go/CLAUDE.md](../go/CLAUDE.md) - Go版の技術スタック、開発環境、コーディング規約
