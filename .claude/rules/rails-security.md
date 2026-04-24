---
paths:
  - "rails/**/*.{rb,erb,js,ts}"
---

# セキュリティガイドライン（Rails 版）

このドキュメントは、Rails 版 Wikino でのセキュリティベストプラクティスを説明します。

## 基本方針

Web アプリケーションのセキュリティは**最優先事項**です。以下のガイドラインを必ず守ってください。

## CSRF（Cross-Site Request Forgery）対策

### Rails 標準の保護

Rails では `protect_from_forgery with: :exception` がデフォルトで有効になっています。Wikino では `ApplicationController` でこれを継承しており、POST / PATCH / DELETE リクエストには自動的に CSRF トークンの検証が行われます。

```ruby
# app/controllers/application_controller.rb
class ApplicationController < ActionController::Base
  # protect_from_forgery は Rails のデフォルトで有効
end
```

### フォームでの使用

`form_with` ヘルパーが自動的に CSRF トークンを追加します。

```erb
<%# ✅ Good: form_with は自動的に CSRF トークンを追加 %>
<%= form_with model: @page, url: pages_path do |f| %>
  <%= f.text_field :title %>
  <%= f.submit t(".submit") %>
<% end %>

<%# ❌ Bad: 手動で form タグを書く（CSRF トークンが含まれない） %>
<form action="/pages" method="post">
  <input type="text" name="page[title]" />
</form>
```

### 内部 API の扱い

Wikino の内部 API（`app/controllers/api/internal/`）は、外部公開用ではなく同一オリジン上の JavaScript からのみ呼び出されます。認証は通常のコントローラーと同じ `require_authentication`（Cookie ベース）を使います。

JavaScript からの HTTP リクエストには `@rails/request.js` を使うことで、CSRF トークンが自動的にリクエストヘッダーに付与されます。

```typescript
// ✅ Good: @rails/request.js を使用（CSRF トークンが自動付与される）
import { post } from "@rails/request.js";

const response = await post("/api/internal/attachments", {
  body: data,
  responseKind: "json",
});
```

なお、S3 署名付き URL 発行のように Rails 側の CSRF 検証を適用できないエンドポイントでは、`require_authentication` と `origin` 検証など、別途ガードを設けます（例: `app/controllers/attachments/presigns/create_controller.rb`）。

## XSS（Cross-Site Scripting）対策

### テンプレートの自動エスケープ

ERB / ViewComponent は自動でエスケープ処理を行います。

```erb
<%# ✅ 自動的にエスケープされる %>
<%= @page.title %>  <%# <script>...</script> は &lt;script&gt; になる %>
```

### `raw` / `html_safe` の注意

`raw` や `html_safe` を使う場合は、データが信頼できるソースからのものであることを確認してください。

```erb
<%# ⚠️ 注意: 信頼できる HTML のみ %>
<%= raw @rendered_markdown %>

<%# ❌ NG: ユーザー入力を直接 raw() で使用しない %>
<%= raw @page.body %>

<%# ✅ Good: html-pipeline 経由でサニタイズされた HTML を使う %>
<%= @page.body_html %>  <%# Markup モデル側で html-pipeline + sanitize により無害化済み %>
```

Wikino では Markdown 本文を `Markup` クラス（`app/models/markup.rb`）で処理し、`html-pipeline` と `sanitize` gem で XSS 対策を施してから表示します。SVG を扱う場合は `SvgSanitizer` を通します。

### Content Security Policy

**現状**: Wikino では CSP は**まだ設定していません**。`config/initializers/content_security_policy.rb` は全行コメントアウトされています。

将来 CSP を導入する際は、以下のような設定を参考にしてください。

```ruby
# config/initializers/content_security_policy.rb
Rails.application.configure do
  config.content_security_policy do |policy|
    policy.default_src :self, :https
    policy.font_src    :self, :https, :data
    policy.img_src     :self, :https, :data
    policy.object_src  :none
    policy.script_src  :self, :https
    policy.style_src   :self, :https, :unsafe_inline
  end

  # importmap / インラインスクリプト用の nonce を発行する
  config.content_security_policy_nonce_generator = ->(request) { request.session.id.to_s }
  config.content_security_policy_nonce_directives = %w[script-src style-src]
end
```

## SQL インジェクション対策

### ActiveRecord のプリペアドステートメント

ActiveRecord は自動的にプリペアドステートメントを使用します。

```ruby
# ✅ Good: 安全
UserRecord.where(email: params[:email])
UserRecord.where("email = ?", params[:email])
UserRecord.where("email = :email", email: params[:email])

# ❌ NG: SQL インジェクションの脆弱性
UserRecord.where("email = '#{params[:email]}'")
```

### `where` 条件のプレースホルダー

必ずプレースホルダー（`?` または名前付き）を使用します。`LIKE` 句を使う場合は `sanitize_sql_like` でメタ文字をエスケープします。

```ruby
# ✅ Good: プレースホルダー + LIKE のメタ文字エスケープ
def search_pages(title)
  PageRecord.where("title LIKE ?", "%#{sanitize_sql_like(title)}%")
end

# ❌ NG: 文字列補間
def search_pages(title)
  PageRecord.where("title LIKE '%#{title}%'")
end
```

## パスワード管理

### `has_secure_password` の使用

Wikino ではパスワードを `UserPasswordRecord` に `has_secure_password` で保管します（`users` と別テーブルに分離しています）。bcrypt が内部的に使われ、ソルトも自動生成されます。

```ruby
# app/records/user_password_record.rb
class UserPasswordRecord < ApplicationRecord
  self.table_name = "user_passwords"

  has_secure_password

  belongs_to :user_record, foreign_key: :user_id
end
```

```ruby
# パスワードの作成
UserPasswordRecord.create!(user_record:, password: "secure_password")

# パスワードの検証
user_password_record.authenticate("secure_password")  # 成功時は self、失敗時は false
```

### 平文パスワードの扱い

平文パスワードをログや外部サービスに送出してはいけません。

```ruby
# ❌ NG: 平文パスワードをログに出力
Rails.logger.info "User password: #{params[:password]}"

# ✅ Good: パスワードはログに出力しない
Rails.logger.info "User sign-in attempt: email=#{params[:email]}"
```

Wikino の `config/initializers/filter_parameter_logging.rb` では、`:passw`, `:secret`, `:token`, `:crypt`, `:salt`, `:otp`, `:cvv`, `:cvc` などを自動的にフィルタリングしています。新しい機密パラメータを追加する場合は、このリストも更新してください。

```ruby
# config/initializers/filter_parameter_logging.rb
Rails.application.config.filter_parameters += [
  :passw, :email, :secret, :token, :_key, :crypt, :salt, :certificate, :otp, :ssn, :cvv, :cvc
]
```

## 認証・認可

Wikino は Devise / Pundit を使わず、独自の認証・認可機構を実装しています。詳細は [@.claude/rules/rails-architecture.md](rails-architecture.md) を参照してください。

### 認証: `ControllerConcerns::Authenticatable`

認証は `ControllerConcerns::Authenticatable` モジュール（`app/controllers/controller_concerns/authenticatable.rb`）で提供される `require_authentication` を `before_action` に設定します。

```ruby
# app/controllers/pages/new_controller.rb
module Pages
  class NewController < ApplicationController
    before_action :require_authentication

    def call
      # current_user_record と current_user が使える
      page_form = PageForm.new(user: current_user)
      # ...
    end
  end
end
```

提供されるメソッド:

| メソッド                 | 用途                                                     |
| ------------------------ | -------------------------------------------------------- |
| `require_authentication` | 未認証の場合は `/sign_in` へリダイレクトする             |
| `sign_in(record)`        | ログイン状態に遷移する（`UserSessionRecord` を受け取る） |
| `sign_out`               | ログアウトし、Cookie を削除する                          |
| `signed_in?`             | 認証状態の判定                                           |
| `current_user_record`    | 認証中の `UserRecord` を返す（未認証なら `nil`）         |
| `current_user`           | 認証中の `User` モデルを返す（未認証なら `nil`）         |

### 認可: `MemberPolicy` / `GuestPolicy`

認可は Pundit の代わりに Wikino 独自の Policy クラス（`app/policies/`）で判定します。`MemberPolicy` はスペースメンバーに対して、`GuestPolicy` は非メンバーに対して権限判定を行います。

```ruby
# app/policies/member_policy.rb
class MemberPolicy
  def can_create_page?
    effective_scopes.include?(Scope::PAGE_WRITE)
  end

  def can_update_space?
    effective_scopes.include?(Scope::SPACE_WRITE)
  end
  # ...
end
```

コントローラーから Policy を呼び出して認可チェックを行います。権限がなければ 404 を返す、または適切なエラーページに遷移させます。

```ruby
# app/controllers/pages/create_controller.rb
module Pages
  class CreateController < ApplicationController
    before_action :require_authentication

    def call
      space_member_record = SpaceMemberRecord.find_by!(user_record: current_user_record!, space_record:)
      policy = MemberPolicy.new(
        space_scopes: space_member_record.space_scopes,
        topic_scopes: topic_member_record.topic_scopes
      )

      unless policy.can_create_page?
        return render_404
      end

      # 作成処理
    end
  end
end
```

### 所有者チェック

認証だけでなく、リソースの所有者であるかもチェックします。たとえば編集中の下書きページは所有者のみが編集できます（`can_update_draft_page?(is_owner:)` のように、`MemberPolicy` が `is_owner` を受け取って判定）。

## Strong Parameters

すべてのコントローラーで Strong Parameters を使用し、許可するパラメータを明示します。

```ruby
# ✅ Good: Strong Parameters
private def page_params
  params.require(:page).permit(:title, :body)
end

def call
  @page = PageRecord.new(page_params)
  # ...
end

# ❌ NG: パラメータを直接使用（Mass Assignment 脆弱性）
def call
  @page = PageRecord.new(params[:page])
end
```

Wikino では多くの入力を Form オブジェクト（`app/forms/`）で受け取り、Form 側で許可パラメータとバリデーションを管理するケースもあります。どちらの場合も「明示的に許可したパラメータのみを使う」という原則は共通です。

## セッション管理

Wikino は 2 種類のセッションを併用しています。

| 用途                   | 実装                                           | 保存先                                                  |
| ---------------------- | ---------------------------------------------- | ------------------------------------------------------- |
| フラッシュメッセージ等 | Rails 標準の CookieStore                       | `_wikino_session` Cookie                                |
| ユーザー認証           | `user_sessions` テーブル + 認証トークン Cookie | `user_sessions` テーブル + `user_session_tokens` Cookie |

### ユーザー認証セッション

ログイン時に `UserSessionRecord.start!` を呼び、`has_secure_token` で生成された**新しいトークン**をレコードに保存します。このトークンを Cookie に保存し、以降のリクエストで認証に使います。

ログインのたびにレコードが新規作成されトークンが更新されるため、**セッション固定攻撃（Session Fixation）は構造的に防がれています**。

```ruby
# app/records/user_session_record.rb
class UserSessionRecord < ApplicationRecord
  self.table_name = "user_sessions"

  has_secure_token  # token カラムにランダム 24 文字を自動生成

  belongs_to :user_record, foreign_key: :user_id

  def self.start!(ip_address:, user_agent:, signed_in_at: Time.current)
    create!(ip_address: ip_address || "", user_agent: user_agent || "", signed_in_at:)
  end
end
```

### Cookie の設定

認証トークン Cookie は `ControllerConcerns::Authenticatable#store_user_session_token` で以下の属性を付与して保存します。

```ruby
cookies.permanent[UserSession::TOKENS_COOKIE_KEY] = {
  value: token,
  httponly: true,              # JavaScript からアクセス不可（XSS 経由の窃取を防ぐ）
  same_site: :lax,             # クロスサイトからの送信を制限（CSRF 耐性を強化）
  domain: ".#{Wikino.config.host}"  # サブドメイン間で共有
}
```

| 属性        | 値               | 目的                                                  |
| ----------- | ---------------- | ----------------------------------------------------- |
| `httponly`  | `true`           | JavaScript からアクセス不可にし、XSS 経由の窃取を防ぐ |
| `same_site` | `:lax`           | クロスサイトからの送信を制限し、CSRF 耐性を上げる     |
| `domain`    | `.<host>`        | サブドメイン間での共有                                |
| `secure`    | (本番で自動付与) | `config.force_ssl = true` により HTTPS 限定送信       |

### ログアウト

`sign_out` は `UserSessions::DestroyService` で `user_sessions` レコードを削除し、Cookie も削除します。これによりサーバー側のトークンが無効化され、万一 Cookie が流出していても再利用できなくなります。

## エラーメッセージ

### 詳細な情報を漏らさない

ユーザーには一般的なエラーメッセージを表示し、詳細な情報（SQL エラー、スタックトレースなど）は漏らさないようにします。

```ruby
# ❌ NG: 詳細なエラーメッセージをユーザーに表示
rescue ActiveRecord::RecordNotFound => e
  render json: {error: e.message, backtrace: e.backtrace}, status: :not_found
end

# ✅ Good: 一般的なエラーメッセージを表示
rescue ActiveRecord::RecordNotFound
  render_404
end
```

本番環境では `config.consider_all_requests_local = false` に設定されており（`config/environments/production.rb`）、Rails の詳細なエラーページが表示されることはありません。

### Sentry による自動エラー追跡

Wikino では `sentry-rails` / `sentry-ruby` gem を使用し、例外を自動的に Sentry へ送信します。`sentry-rails` は Rack ミドルウェアとして機能するため、コントローラー内で明示的な `rescue_from` を書く必要はありません。

```ruby
# config/initializers/sentry.rb
Sentry.init do |config|
  config.dsn = Wikino.config.sentry_dsn
  config.breadcrumbs_logger = %i[active_support_logger http_logger]
  config.traces_sample_rate = 0.5
  config.profiles_sample_rate = 0.5
end
```

任意の箇所で明示的にエラーを送りたい場合のみ `Sentry.capture_exception(exception)` を呼びます。

## セキュリティヘッダー

### Rails デフォルトのヘッダー

`config.load_defaults 8.0`（`config/application.rb`）により、以下のセキュリティヘッダーが自動設定されます。

| ヘッダー                 | 値                                | 目的                                            |
| ------------------------ | --------------------------------- | ----------------------------------------------- |
| `X-Frame-Options`        | `SAMEORIGIN`                      | クリックジャッキング対策                        |
| `X-Content-Type-Options` | `nosniff`                         | MIME タイプスニッフィング対策                   |
| `X-XSS-Protection`       | `0`                               | 旧ブラウザの XSS Auditor を無効化（誤検知防止） |
| `Referrer-Policy`        | `strict-origin-when-cross-origin` | リファラーの漏洩を抑制                          |

### HSTS（本番環境）

`config/environments/production.rb` で `config.force_ssl = true` を設定しており、HSTS ヘッダーが自動で付与されます。

```ruby
# config/environments/production.rb
Rails.application.configure do
  config.force_ssl = true  # HTTPS を強制、HSTS ヘッダーを自動付与
end
```

### Permissions Policy

**現状**: Wikino では Permissions Policy は**まだ設定していません**。`config/initializers/permissions_policy.rb` は全行コメントアウトされています。

カメラ・マイク・位置情報などの機能を制限したい場合は、以下のような設定を参考にしてください。

```ruby
# config/initializers/permissions_policy.rb
Rails.application.config.permissions_policy do |policy|
  policy.camera     :none
  policy.microphone :none
  policy.geolocation :none
  policy.usb        :none
end
```

## セキュリティチェックリスト

新機能を実装する際は、以下を必ず確認してください。

### フォーム送信

- [ ] `form_with` ヘルパーを使用しているか
- [ ] CSRF トークンが含まれているか
- [ ] JavaScript からのリクエストは `@rails/request.js` を使っているか

### ユーザー入力

- [ ] Strong Parameters または Form オブジェクトで許可パラメータを明示しているか
- [ ] バリデーションを実施しているか
- [ ] ホワイトリスト方式で許可しているか

### データベース

- [ ] ActiveRecord を使用しているか
- [ ] プレースホルダーを使用しているか
- [ ] 文字列補間を避けているか
- [ ] `LIKE` 句では `sanitize_sql_like` を使っているか

### パスワード

- [ ] `has_secure_password` を使用しているか（`UserPasswordRecord`）
- [ ] 平文パスワードをログに出力していないか
- [ ] `filter_parameter_logging` に機密パラメータが登録されているか

### 認証・認可

- [ ] `before_action :require_authentication` を設定しているか
- [ ] `MemberPolicy` / `GuestPolicy` で認可チェックを行っているか
- [ ] リソースの所有者チェックを行っているか

### エラー処理

- [ ] エラーメッセージは適切か（詳細な情報を漏らしていないか）

### HTML 出力

- [ ] `raw` / `html_safe` を使う箇所は信頼できるソースか
- [ ] ユーザー入力を含む HTML は `html-pipeline` や `sanitize` を経由しているか

## ベストプラクティス

### 1. bundler-audit で脆弱性をチェック

**現状**: Wikino では bundler-audit は**まだ導入していません**。依存ライブラリの CVE を定期的にチェックしたい場合、導入を検討してください。

```sh
# インストール後、以下のコマンドで実行
bundle audit check --update
```

### 2. Brakeman で静的解析

**現状**: Wikino では Brakeman は**まだ導入していません**。Rails 固有のセキュリティ脆弱性を静的に検出したい場合、導入を検討してください。

```sh
# インストール後、以下のコマンドで実行
brakeman -q
```

### 3. 環境変数でシークレットを管理

シークレット（API キー、DSN など）はソースコードに直接書かず、環境変数経由で読み込みます。

```ruby
# ❌ NG: ハードコード
API_KEY = "abc123"

# ✅ Good: 環境変数
API_KEY = ENV.fetch("WIKINO_API_KEY")
```

Wikino では `.env.{environment}` ファイルで環境変数を管理します。シークレットを含む `.env.local` は Git 管理外です。

### 4. 定期的に Gem を更新

Dependabot が有効になっており、依存 Gem の更新 PR が自動で作成されます。PR 単位でセキュリティ更新を取り込むため、マージは定期的に行ってください。

## トラブルシューティング

### CSRF トークンエラー

**症状**: `ActionController::InvalidAuthenticityToken`

**原因**:

1. フォームに CSRF トークンが含まれていない（手書きの `<form>` タグを使っている）
2. セッションが切れている
3. JavaScript リクエストにトークンが含まれていない

**解決方法**:

```erb
<%# フォーム: form_with を使用 %>
<%= form_with model: @page do |f| %>
  <%= f.submit %>
<% end %>
```

```typescript
// JavaScript: @rails/request.js を使用
import { post } from "@rails/request.js";

const response = await post("/api/internal/...", {
  body: data,
  responseKind: "json",
});
```

### Mass Assignment

**症状**: 想定外のパラメータが保存されてしまう

**原因**: Strong Parameters または Form オブジェクトでパラメータを絞り込んでいない

**解決方法**:

```ruby
# ✅ Good: Strong Parameters
private def page_params
  params.require(:page).permit(:title, :body)
end
```

## 参考資料

- [Rails Security Guide](https://guides.rubyonrails.org/security.html)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Brakeman](https://brakemanscanner.org/)
- [Bundler Audit](https://github.com/rubysec/bundler-audit)
