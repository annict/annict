---
paths:
  - "rails/**"
---

# Wikino 開発ガイド (Rails 版)

このファイルは、Rails 版 Wikino の開発に関するガイダンスを提供します。

> **Note**: プロジェクト全体の概要、モノレポ構造、共通インフラ（PostgreSQL）については、[/CLAUDE.md](../CLAUDE.md) を参照してください。

## Rails 版の開発方針

Rails 版は既存の本番システムであり、以下の方針で開発・保守を進めています：

- **安定性優先**: 既存機能の動作を維持しながら慎重に改善
- **段階的な Go 版移行**: 機能ごとに段階的に Go 版へ移行
- **共通インフラの継続利用**: PostgreSQL などは Go 版と共有

### Go 版への移行について

現在、Rails 版の機能を段階的に Go 版へ移行中です。Go 版の実装時には、Rails 版のコードを参考にして既存の仕様を理解できます。

## プロジェクト概要

WikinoはWikiアプリケーションです。
ユーザーは「スペース」と呼ばれる場所にページを作成し、ページ間をリンクで繋げることができます。

## 技術スタック

### バックエンド

- Ruby 3.4.4
- Ruby on Rails 8.0.0
- PostgreSQL
- Sorbet（型検査）
- Active Job（Solid Queue）

### フロントエンド

- TypeScript
- Hotwire (Turbo + Stimulus)
- Tailwind CSS 4
- CodeMirror 6（ページエディタ）

### ツール・ライブラリ

- パッケージマネージャー: Bundler, pnpm
- テスティング: RSpec, FactoryBot
- Linter: Standard (Ruby), ERB Lint, ESLint
- Formatter: Oxfmt（プロジェクトルートで管理）
- ViewComponent, html-pipeline, meta-tags

## プロジェクト構造

### app/ディレクトリの構成と責務

| ディレクトリ      | 責務                   | 説明                                              |
| ----------------- | ---------------------- | ------------------------------------------------- |
| **controllers/**  | HTTPリクエスト処理     | 1アクション1コントローラー、`#call`メソッドで実装 |
| **records/**      | DBテーブル操作         | ActiveRecord::Base継承、1テーブル1レコード        |
| **models/**       | ドメインロジック       | PORO、データベースアクセスなし                    |
| **repositories/** | データ変換             | RecordとModel間の変換                             |
| **services/**     | ビジネスロジック       | **データ永続化を伴う処理のみ**実装                |
| **forms/**        | フォーム処理           | バリデーションとデータ変換                        |
| **components/**   | UIコンポーネント       | ViewComponent、再利用可能なUI要素                 |
| **views/**        | ビュー                 | ViewComponent使用、DB直接アクセス禁止             |
| **policies/**     | 認可ルール             | 権限管理                                          |
| **validators/**   | カスタムバリデーション | ActiveModelバリデーター拡張                       |
| **jobs/**         | 非同期処理             | 最小限のロジック、主にService呼び出し             |
| **mailers/**      | メール送信             | Action Mailer                                     |

## Railsクラス設計と依存関係

📖 **詳細については [@.claude/rules/rails-architecture.md](rails-architecture.md) を参照してください。**

### クラス間の依存関係ルール

| クラス     | 依存可能な先                                   |
| ---------- | ---------------------------------------------- |
| Component  | Component, Form, Model                         |
| Controller | Form, Model, Record, Repository, Service, View |
| Form       | Record, Validator                              |
| Job        | Service                                        |
| Mailer     | Model, Record, Repository, View                |
| Model      | Model                                          |
| Policy     | Record                                         |
| Record     | Record                                         |
| Repository | Model, Record, Policy                          |
| Service    | Job, Mailer, Record                            |
| Validator  | Record                                         |
| View       | Component, Form, Model                         |

### 命名規則

- Controller: `(ModelPlural)::(ActionName)Controller`
- Service: `(ModelPlural)::(Verb)Service`
- Form: `(ModelPlural)::(Noun)Form`
- Repository: `(Model)Repository`
- View: `(ModelPlural)::(ActionName)View`
- Component: `(UIComponentPlural)::(Noun)Component`

## コーディング規約（プロジェクト固有）

### Ruby

```ruby
# typed: strict
# frozen_string_literal: true

class Example
  # ✅ 文字列はダブルクオート
  name = "example"

  # ✅ ハッシュの省略記法
  {user:, name:}

  # ❌
  {user: user, name: name}

  # ✅ プライベートメソッドは private def
  private def process_value(value)
    value.upcase
  end

  # ✅ プロテクテッドメソッドは protected def
  protected def shared_method(value)
    value.downcase
  end

  # ❌ 後置ifは使用しない
  # return if value.nil? # 悪い例

  # ✅
  if value.nil?
    return
  end

  # ❌ attr_readerにprivateブロックを使用しない
  # private
  # attr_reader :user_record

  # ✅ attr_readerは個別にprivate指定
  attr_reader :user_record
  private :user_record

  # ✅ T.mustではなくnot_nil!を使用
  value.not_nil!
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

### ActiveRecord

```ruby
# ❌ includesは使用禁止
Model.includes(:association)

# ✅ 明示的にpreloadまたはeager_loadを使用
Model.preload(:association)   # 別クエリで取得（基本はこちら）
Model.eager_load(:association) # JOINで取得（関連テーブルでフィルタリング時）
```

### マイグレーション

```ruby
create_table :examples, id: false do |t|
  # ULIDを使用
  t.uuid :id, default: "generate_ulid()", null: false, primary_key: true
end
```

### 型定義

```ruby
# データベースIDの型はTypes::DatabaseIdを使用
sig { params(space_record_id: Types::DatabaseId).returns(T::Boolean) }
def in_same_space?(space_record_id:)
  # ...
end

# ❌ 単純なStringではなく
sig { params(space_record_id: String).returns(T::Boolean) }

# ✅ Types::DatabaseIdを使用
sig { params(space_record_id: Types::DatabaseId).returns(T::Boolean) }
```

### RSpec

📖 **詳細については [@.claude/rules/rails-testing.md](rails-testing.md) を参照してください。**

```ruby
# ❌ context, let, described_classは使用しない
context "when xxx" do
  let(:user) { create(:user) }
end

# ✅ itブロック内で変数定義
it "xxxのとき、somethingすること" do
  user = FactoryBot.create(:user)
  # テスト実装
end

# ✅ FactoryBotで作成したレコードの変数名には_recordサフィックスを付ける
user_record = FactoryBot.create(:user_record)
space_record = FactoryBot.create(:space_record)
space_member_record = FactoryBot.create(:space_member_record, user_record:, space_record:)

# ❌ サフィックスなしの変数名は避ける
user = FactoryBot.create(:user_record)
space = FactoryBot.create(:space_record)
```

#### システムテストの待機処理

```ruby
# ❌ sleepを使用した待機処理は避ける
button.click
sleep 2
expect(page).to have_current_path(some_path)

# ✅ Capybaraの待機機能を活用
button.click
# ページ上の要素の変化を待つ（Capybaraが自動的に最大5秒待機）
expect(page).not_to have_content("削除されたコンテンツ")
expect(page).to have_content("新しく表示されるコンテンツ")

# ✅ have_css/not_to have_cssで要素の出現/消失を待つ
expect(page).to have_css(".success-message")
expect(page).not_to have_css(".loading-spinner")
```

**重要**: システムテストでは`sleep`の使用を避け、Capybaraの自動待機機能を活用すること。要素の出現や消失、コンテンツの変化を検証することで、適切な待機処理が自動的に行われる。

### I18n

翻訳ファイルは用途別に分類し、日本語と英語の両方を更新：

- `forms.(ja,en).yml`: フォーム関連
- `messages.(ja,en).yml`: メッセージ・説明文
- `meta.(ja,en).yml`: メタデータ
- `nouns.(ja,en).yml`: 名詞・ラベル

### JavaScript/TypeScript

#### HTTPリクエスト

```typescript
// ❌ fetchを使用しない
const response = await fetch("/api/endpoint", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-CSRF-Token": csrfToken,
  },
  body: JSON.stringify(data),
});

// ✅ @rails/request.jsを使用
import { post } from "@rails/request.js";

const response = await post("/api/endpoint", {
  body: data,
  responseKind: "json",
});
```

**重要**: Railsアプリケーション内でのHTTPリクエストには、`fetch`ではなく`@rails/request.js`パッケージを使用すること。CSRFトークンの管理が自動化され、Railsとの統合がよりシームレスになります。

## サービスクラスのルール

📖 **詳細については [@.claude/rules/rails-architecture.md](rails-architecture.md) を参照してください。**

### サービスクラスを使用する場合

- ✅ データベースへの永続化を伴う処理
- ✅ 複数のモデル/レコードにまたがる複雑なビジネスロジックで永続化を伴うもの
- ✅ トランザクション管理が必要な処理

### サービスクラスを使用しない場合

- ❌ データベースへの永続化を伴わない処理（URL生成、データ変換など）
- ❌ 単一のモデル/レコードに閉じた処理（モデルやレコードのメソッドとして定義）

### トランザクション処理

**重要**: Serviceクラスでトランザクションを張る場合は、必ず `#with_transaction` メソッドを使用すること

```ruby
# ✅ 良い例：with_transactionを使用
module Users
  class CreateService < ApplicationService
    def call
      with_transaction do
        user = UserRecord.create!(...)
        ProfileRecord.create!(user:, ...)
      end
    end
  end
end

# ❌ 悪い例：transactionを直接使用
module Users
  class CreateService < ApplicationService
    def call
      ApplicationRecord.transaction do
        # with_transactionを使うべき
      end
    end
  end
end
```

**重要**: Controller、Job、Rakeタスク内で永続化処理を書く場合は、必ずServiceクラスを定義すること

## 開発環境のセットアップ

> **Note**: 開発環境の基本的なセットアップ手順は [/CLAUDE.md](../CLAUDE.md#開発環境のセットアップ) を参照してください。

- Claude Code はコンテナ内で実行されているため、ホスト側のコマンドの実行は不要です
- 共通インフラ（PostgreSQL）は `/workspace/docker-compose.yml` で管理されており、ホスト側で起動済みのはずです

### 環境変数の設定

環境変数は`.env.{environment}`ファイルで管理します：

- `.env.development` - 開発環境用
- `.env.test` - テスト環境用
- `.env.local` - ローカル固有の設定（機密情報など、Git管理外）

### ホスト側で実行するコマンド (Claude Code による実行は不要)

```sh
# コンテナ起動
docker compose up

# 特定のサービスのログを確認
docker compose logs -f app
```

### コンテナ内で実行するコマンド (Claude Code が実行できるコマンド)

**重要**: 環境変数は 1Password CLI 経由で読み込まれるため、`bin/rails` などを直接実行せず、`make` コマンドを使用してください。

```bash
# セットアップ（依存関係のインストール、DB作成など）
make setup

# データベースのセットアップ
make db-setup

# 開発サーバー起動
make server

# コンソール起動
make console

# テスト実行
make test

# 特定のテストを実行
make test-file FILE=spec/path/to/spec.rb

# コードフォーマット
make fmt                  # Ruby（自動修正）
make -C /workspace fmt    # Oxfmt（JS/TS/CSS/YAML/Markdown/TOML/JSON）

# リント
make lint                 # Ruby
bin/erb_lint --lint-all   # ERB
pnpm eslint . --fix       # JavaScript/TypeScript

# Sorbet型チェック
make sorbet

# Sorbet型定義を更新
make sorbet-update

# Zeitwerk（オートロード）チェック
make zeitwerk

# PostgreSQLに接続
psql $DATABASE_URL

# データベースマイグレーション
make db-migrate
make db-rollback          # 最後のマイグレーションをロールバック

# GraphQL APIスキーマをダンプ
make graphql-dump
```

### コミット前に実行するコマンド

**重要**: コードをコミットする前に、以下のコマンドを実行して CI が通ることを確認してください：

```bash
# 1. Zeitwerk（オートロード）チェック
make zeitwerk

# 2. Sorbet型定義の更新と型チェック
make sorbet-update
make sorbet

# 3. Rubyコードのリント・フォーマット
make fmt

# 4. ERBリント
bin/erb_lint --lint-all

# 5. Oxfmt（JS/TS/CSS/YAML/Markdown/TOML/JSON）
make -C /workspace fmt

# 6. ESLint
pnpm eslint . --fix

# 7. TypeScript型チェック
pnpm tsc

# 8. テストを実行
make test

# すべてを一度に実行するワンライナー:
make zeitwerk && make sorbet-update && make sorbet && make fmt && bin/erb_lint --lint-all && make -C /workspace fmt && pnpm eslint . --fix && pnpm tsc && make test
```

## Pull Request のガイドライン

Pull Request のガイドラインは [/CLAUDE.md](../CLAUDE.md#pull-requestのガイドライン) を参照してください。

**要約**:

- 実装コード: 300 行以下を目安
- テストコード: 制限なし（必要な分だけ書く）
- 実装とテストは同じ PR に含める
- 「行数を守ること」よりも「きちんと実装すること」を優先

## 作業完了ガイドライン

### タスク実装フロー

#### 1. タスク理解

- 要件を理解
- このガイドの固有ルールを確認

#### 2. 実装前の準備

- 既存コードの調査
- 特に以下を意識：
  - プライベートメソッドは `private def`
  - `includes` ではなく `preload` / `eager_load`
  - `T.must` ではなく `not_nil!`

#### 3. 実装

- 規約に従ってコーディング
- 新規ファイルにはマジックコメント追加

#### 4. 完了前の検証

**重要**: 完了報告前に全ての作業が適切に検証されていることを確認すること

- **テスト作成**: テスト作成後は、必ず `make test` を実行してテストが通ることを確認する
- **コード実装**: コード記述後は、必ず以下を確認する:
  - 型チェックが成功すること（`make sorbet` or `pnpm tsc`）
  - Linterの実行が成功すること（`make lint` or `bin/erb_lint --lint-all` or `pnpm eslint . --fix`）
  - 関連するテストが通ること（`make test-file FILE=spec/path/to/xxx_spec.rb`）
  - 明らかなランタイムエラーがないこと
- **ドキュメント編集**: Markdownファイルを編集した後は、必ず以下を行う:
  - Oxfmtの実行 (`make -C /workspace fmt`)
- **リトライポリシー**: 問題発生時は自動で最大5回まで再試行し、それでも解消できない場合にのみユーザーへ連絡する（途中経過は報告しない）
  - Report to user: "同じエラーが5回続いています。別のアプローチが必要かもしれません。"
- **以下の状態では絶対に完了報告をしない**:
  - テストが失敗している（未実装機能のテストを意図的に作成している場合を除く）
  - コンパイルエラーがある
  - 前回の試行から未解決のエラーが残っている

### 検証コマンド

```bash
# Rubyのファイルを編集したとき実行する
make lint
bin/erb_lint --lint-all
make sorbet
make sorbet-update
make zeitwerk
make test

# JavaScript/TypeScriptを編集したとき実行する
make -C /workspace fmt
pnpm eslint . --fix
pnpm tsc
```

## デバッグ・トラブルシューティング

- Sorbetエラー: `make sorbet-update` で型定義更新
- オートローディングエラー: `make zeitwerk`
- フォーマットエラー: `make -C /workspace fmt`
- Lintエラー: 各種Lintコマンドで修正

## セキュリティガイドライン

📖 **詳細については [@.claude/rules/rails-security.md](rails-security.md) を参照してください。**

Web アプリケーションのセキュリティは**最優先事項**です。

### 基本対策

- **CSRF 対策**: `protect_from_forgery` がデフォルトで有効、`form_with` ヘルパーを使用
- **XSS 対策**: ERB の自動エスケープを活用、`raw`/`html_safe` は慎重に使用
- **SQL インジェクション対策**: ActiveRecord のプリペアドステートメント、プレースホルダーを使用
- **認証**: bcrypt（`has_secure_password`）で管理
- **Strong Parameters**: すべてのコントローラーで使用

## 重要な原則

- 説明的な命名規則
- コメントは日本語で記載
- 1行100文字以内
- セキュリティベストプラクティスに従う
