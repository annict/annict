# Rails 認証ページコードの削除 設計書

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

Go 版でログインとユーザー登録ページの実装が完了したため、Rails 側の対応するコードを削除する。
これにより、コードベースの重複を排除し、保守性を向上させる。

**目的**:

- Go 版への認証機能移行を完了させる
- Rails 側の不要なコードを削除し、コードベースをシンプルに保つ
- 将来的なメンテナンスコストを削減

**背景**:

- Go 版でログイン、サインアップ、パスワードリセット機能が実装済み
- Rails 版と Go 版で認証関連コードが重複している
- リバースプロキシにより、`/sign_in`, `/sign_up`, `/password/*` などは Go 版にルーティングされている

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

- Rails 側のログイン関連コード（コントローラー、ビュー、フォーム、テスト）を削除する
- Rails 側のサインアップ関連コード（コントローラー、ビュー、フォーム、テスト）を削除する
- Rails 側のパスワードログイン（レガシー）関連コードを削除する
- ルーティング設定から削除したエンドポイントを除去する
- 関連する国際化メッセージを削除する
- 関連するメールテンプレート（sign_in/sign_up 確認メール）を削除する

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

- **後方互換性**: 既存のログインセッションに影響を与えない
- **保守性**: 削除後も Rails アプリケーションが正常に動作すること
- **段階的削除**: 1 PR で削除する量を適切に管理し、問題が発生した場合のロールバックを容易にする

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

### 削除対象ファイル一覧

#### コントローラー（約 228 行）

| ファイル                                                   | 行数 | 役割                             |
| ---------------------------------------------------------- | ---- | -------------------------------- |
| `app/controllers/sign_in_controller.rb`                    | -    | ログインフォーム表示             |
| `app/controllers/sign_up_controller.rb`                    | -    | サインアップフォーム表示         |
| `app/controllers/registrations_controller.rb`              | -    | ユーザー登録確認フォーム表示     |
| `app/controllers/sign_in_callbacks_controller.rb`          | -    | ログインメールのコールバック処理 |
| `app/controllers/legacy/sessions_controller.rb`            | -    | レガシーパスワードログイン       |
| `app/controllers/api/internal/sign_in_controller.rb`       | -    | ログイン API                     |
| `app/controllers/api/internal/sign_up_controller.rb`       | -    | サインアップ API                 |
| `app/controllers/api/internal/registrations_controller.rb` | -    | ユーザー登録 API                 |

#### フォームオブジェクト（約 75 行）

| ファイル                                | 行数 | 役割                               |
| --------------------------------------- | ---- | ---------------------------------- |
| `app/models/forms/sign_in_form.rb`      | -    | ログインフォームバリデーション     |
| `app/models/forms/sign_up_form.rb`      | -    | サインアップフォームバリデーション |
| `app/models/forms/registration_form.rb` | -    | ユーザー登録フォームバリデーション |

#### ビューテンプレート（約 303 行）

| ファイル                                    | 行数 | 役割                     |
| ------------------------------------------- | ---- | ------------------------ |
| `app/views/sign_in/new.html.erb`            | -    | ログインフォーム         |
| `app/views/sign_up/new.html.erb`            | -    | サインアップフォーム     |
| `app/views/sign_in_callbacks/show.html.erb` | -    | ログインコールバック     |
| `app/views/registrations/new.html.erb`      | -    | ユーザー登録フォーム     |
| `app/views/legacy/sessions/new.html.erb`    | -    | レガシーログインフォーム |

#### メールテンプレート（約 30 行）

| ファイル                                                               | 行数 | 役割                             |
| ---------------------------------------------------------------------- | ---- | -------------------------------- |
| `app/views/email_confirmation_mailer/sign_in_confirmation.ja.html.erb` | -    | ログイン確認メール（日本語）     |
| `app/views/email_confirmation_mailer/sign_in_confirmation.en.html.erb` | -    | ログイン確認メール（英語）       |
| `app/views/email_confirmation_mailer/sign_up_confirmation.ja.html.erb` | -    | サインアップ確認メール（日本語） |
| `app/views/email_confirmation_mailer/sign_up_confirmation.en.html.erb` | -    | サインアップ確認メール（英語）   |

#### テストファイル（約 826 行）

| ファイル                                                  | 行数 | 役割                                     |
| --------------------------------------------------------- | ---- | ---------------------------------------- |
| `spec/requests/sign_in/new_spec.rb`                       | 55   | ログインフォーム表示テスト               |
| `spec/requests/sign_up/new_spec.rb`                       | 32   | サインアップフォーム表示テスト           |
| `spec/requests/registrations/new_spec.rb`                 | 35   | ユーザー登録フォーム表示テスト           |
| `spec/requests/sign_in_callbacks/show_spec.rb`            | 66   | ログインコールバックテスト               |
| `spec/requests/api/internal/sign_in/create_spec.rb`       | 115  | ログイン API テスト                      |
| `spec/requests/api/internal/sign_up/create_spec.rb`       | 107  | サインアップ API テスト                  |
| `spec/requests/api/internal/registrations/create_spec.rb` | 184  | ユーザー登録 API テスト                  |
| `spec/models/forms/sign_in_form_spec.rb`                  | 46   | ログインフォームバリデーションテスト     |
| `spec/models/forms/sign_up_form_spec.rb`                  | 46   | サインアップフォームバリデーションテスト |
| `spec/models/forms/registration_form_spec.rb`             | 140  | ユーザー登録フォームバリデーションテスト |

#### ルーティング（`config/routes.rb` の修正）

削除対象のルート:

```ruby
# 削除対象
match "/sign_in",                 via: :get,    as: :sign_in,                 to: "sign_in#new"
match "/sign_in",                 via: :get,    as: :new_user_session,        to: "sign_in#new"
match "/sign_in/callback",        via: :get,    as: :sign_in_callback,        to: "sign_in_callbacks#show"
match "/sign_up",                 via: :get,    as: :sign_up,                 to: "sign_up#new"
match "/registrations/new",       via: :get,    as: :new_registration,        to: "registrations#new"
match "/api/internal/sign_in",    via: :post,   as: :internal_api_sign_in,    to: "api/internal/sign_in#create"
match "/api/internal/sign_up",    via: :post,   as: :internal_api_sign_up,    to: "api/internal/sign_up#create"
match "/api/internal/registrations", via: :post, as: :internal_api_registrations, to: "api/internal/registrations#create"

# devise_scope 内
match "/legacy/sign_in",          via: :get,    as: :legacy_sign_in,          to: "legacy/sessions#new"
match "/legacy/sign_in",          via: :post,   as: :user_session,            to: "legacy/sessions#create"
```

#### メーラー（`app/mailers/email_confirmation_mailer.rb` の修正）

削除対象のメソッド:

```ruby
def sign_up_confirmation(email_confirmation_id, locale)
def sign_in_confirmation(email_confirmation_id, locale)
```

保持するメソッド:

```ruby
def update_email_confirmation(email_confirmation_id, locale)  # メール更新確認用、削除しない
```

#### Devise メールテンプレート（削除対象）

| ファイル                                                          | 行数 | 役割                               |
| ----------------------------------------------------------------- | ---- | ---------------------------------- |
| `app/views/devise/mailer/reset_password_instructions.ja.html.erb` | -    | パスワードリセットメール（日本語） |
| `app/views/devise/mailer/reset_password_instructions.en.html.erb` | -    | パスワードリセットメール（英語）   |

**注**: `config/routes.rb` で `skip: %i[passwords ...]` により Devise のパスワードリセットルートが無効化されており、これらのテンプレートはどこからも使用されていない。Go 版でパスワードリセット機能が実装済みのため、削除可能。

### 保持するファイル

以下のファイルは削除**しない**（パスワード変更、メール更新など別の機能で使用）:

- `app/controllers/settings/passwords_controller.rb` - パスワード変更（ログイン後に現在のパスワードを入力して変更）
- `app/views/settings/passwords/show.html.erb` - パスワード変更フォーム
- `app/views/email_confirmation_mailer/update_email_confirmation.*.html.erb` - メール更新確認メール
- `config/initializers/devise.rb` - Devise 設定（パスワード認証、OAuth で必要）
- `app/models/user.rb` - Devise 設定（パスワード認証で必要）

### 実装方針

1. **段階的な削除**: ファイル数と行数を考慮し、複数の PR に分割して削除
2. **テストの確認**: 各フェーズで `bundle exec rspec` を実行し、削除によるテスト失敗がないことを確認
3. **ルーティングの整理**: コントローラー削除後にルーティングを更新

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

### フェーズ 1: ログイン関連コードの削除

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: ログインフォーム関連ファイルの削除
  - `app/controllers/sign_in_controller.rb` の削除
  - `app/controllers/sign_in_callbacks_controller.rb` の削除
  - `app/controllers/api/internal/sign_in_controller.rb` の削除
  - `app/models/forms/sign_in_form.rb` の削除
  - `app/views/sign_in/` ディレクトリの削除
  - `app/views/sign_in_callbacks/` ディレクトリの削除
  - `app/views/email_confirmation_mailer/sign_in_confirmation.*.html.erb` の削除
  - `spec/requests/sign_in/` ディレクトリの削除
  - `spec/requests/sign_in_callbacks/` ディレクトリの削除
  - `spec/requests/api/internal/sign_in/` ディレクトリの削除
  - `spec/models/forms/sign_in_form_spec.rb` の削除
  - `config/routes.rb` からログイン関連ルートを削除
  - **想定ファイル数**: 約 13 ファイル
  - **想定行数**: 約 300 行削除

### フェーズ 2: サインアップ関連コードの削除

- [x] **2-1**: サインアップフォーム関連ファイルの削除
  - `app/controllers/sign_up_controller.rb` の削除
  - `app/controllers/registrations_controller.rb` の削除
  - `app/controllers/api/internal/sign_up_controller.rb` の削除
  - `app/controllers/api/internal/registrations_controller.rb` の削除
  - `app/models/forms/sign_up_form.rb` の削除
  - `app/models/forms/registration_form.rb` の削除
  - `app/views/sign_up/` ディレクトリの削除
  - `app/views/registrations/` ディレクトリの削除
  - `app/views/email_confirmation_mailer/sign_up_confirmation.*.html.erb` の削除
  - `spec/requests/sign_up/` ディレクトリの削除
  - `spec/requests/registrations/` ディレクトリの削除
  - `spec/requests/api/internal/sign_up/` ディレクトリの削除
  - `spec/requests/api/internal/registrations/` ディレクトリの削除
  - `spec/models/forms/sign_up_form_spec.rb` の削除
  - `spec/models/forms/registration_form_spec.rb` の削除
  - `config/routes.rb` からサインアップ関連ルートを削除
  - **想定ファイル数**: 約 17 ファイル
  - **想定行数**: 約 550 行削除

### フェーズ 3: レガシー認証とメーラーの整理

- [x] **3-1**: レガシーパスワードログインの削除
  - `app/controllers/legacy/sessions_controller.rb` の削除
  - `app/views/legacy/sessions/` ディレクトリの削除
  - `config/routes.rb` からレガシーログイン関連ルートを削除
  - **想定ファイル数**: 約 3 ファイル
  - **想定行数**: 約 60 行削除

- [x] **3-2**: EmailConfirmationMailer と Devise メールテンプレートの整理
  - `app/mailers/email_confirmation_mailer.rb` から `sign_up_confirmation` と `sign_in_confirmation` メソッドを削除
  - `app/views/devise/mailer/reset_password_instructions.*.html.erb` の削除（Devise のパスワードリセットルートは無効化済み、Go 版で実装済み）
  - 関連するテストがあれば削除
  - **想定ファイル数**: 約 3-4 ファイル
  - **想定行数**: 約 50 行削除

### フェーズ 4: 国際化メッセージの整理

- [x] **4-1**: 不要な国際化メッセージの削除
  - `config/locales/messages.ja.yml` から `sign_in`, `sign_up`, `registrations`, `sign_in_callback` 関連のメッセージを削除
  - `config/locales/messages.en.yml` から同様のメッセージを削除
  - **想定ファイル数**: 約 2 ファイル
  - **想定行数**: 約 50 行削除

### フェーズ 5: Recaptcha 関連の削除

- [x] **5-1**: Recaptcha 関連ファイルの削除
  - `app/models/recaptcha.rb` の削除
  - `app/views/application/_recaptcha.html.erb` の削除
  - `app/components/deprecated/inputs/recaptcha_input_component.rb` の削除
  - `config/locales/messages.ja.yml` から `recaptcha` 関連のメッセージを削除
  - `config/locales/messages.en.yml` から同様のメッセージを削除
  - `.standard_todo.yml` から Recaptcha 関連のエントリを削除（存在する場合）
  - **想定ファイル数**: 約 5 ファイル
  - **想定行数**: 約 80 行削除
  - **備考**: Go 版では Cloudflare Turnstile を使用しているため、Recaptcha は不要

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**削除しません**：

- **パスワード変更機能**: `settings/passwords_controller.rb` と関連ファイル（ログイン後の設定画面で使用）
- **メール更新確認機能**: `update_email_confirmation` メソッドと関連テンプレート（メールアドレス変更時に使用）
- **Devise 設定全体**: `config/initializers/devise.rb`（パスワード認証、OAuth など他の機能で必要）
- **OAuth 認証**: `app/controllers/callbacks_controller.rb`（Facebook/Gumroad OAuth で使用）
- **ログアウト機能**: `devise/sessions#destroy`（セッション破棄は Rails 側で継続して処理）

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Go 版実装設計書](./../go.md)
- [Rails Devise ドキュメント](https://github.com/heartcombo/devise)
