# セキュリティガイドライン（Rails版）

このドキュメントは、Rails版Annictでのセキュリティベストプラクティスを説明します。

## 基本方針

Web アプリケーションのセキュリティは**最優先事項**です。以下のガイドラインを必ず守ってください。

## CSRF（Cross-Site Request Forgery）対策

### Rails標準の保護

Rails では`protect_from_forgery with: :exception` がデフォルトで有効になっています。

```ruby
# app/controllers/application_controller.rb
class ApplicationController < ActionController::Base
  protect_from_forgery with: :exception
end
```

### フォームでの使用

`form_with` ヘルパーが自動的にCSRFトークンを追加します。

```ruby
# ✅ Good: form_withは自動的にCSRFトークンを追加
<%= form_with model: @work, url: works_path do |f| %>
  <%= f.text_field :title %>
  <%= f.submit "作成" %>
<% end %>

# ❌ Bad: 手動でformタグを書く（CSRFトークンが含まれない）
<form action="/works" method="post">
  <input type="text" name="work[title]" />
</form>
```

### API でのトークン認証

GraphQLやREST APIではトークン認証を使用します。

```ruby
# app/controllers/api/base_controller.rb
module Api
  class BaseController < ActionController::API
    protect_from_forgery with: :null_session

    before_action :authenticate_with_token

    private

    def authenticate_with_token
      token = request.headers['Authorization']&.gsub(/^Bearer /, '')
      @current_user = User.find_by_access_token(token)
      render json: { error: 'Unauthorized' }, status: :unauthorized unless @current_user
    end
  end
end
```

## XSS（Cross-Site Scripting）対策

### テンプレートの自動エスケープ

ERB/Slimは自動でエスケープ処理を行います。

```ruby
# ✅ 自動的にエスケープされる
<%= @user.comment %>  # <script>...</script> は &lt;script&gt; になる
```

### raw/html_safeの注意

`raw` や `html_safe` を使う場合は、データが信頼できるソースからのものであることを確認してください。

```ruby
# ⚠️ 注意: 信頼できるHTMLのみ
<%= raw @trusted_html_content %>

# ❌ NG: ユーザー入力を直接raw()で使用しない
<%= raw @user.comment %>

# ✅ Good: サニタイズを使用
<%= sanitize @user.comment, tags: %w[p br strong em] %>
```

### Content Security Policy

`config/initializers/content_security_policy.rb` でCSPを設定します。

```ruby
# config/initializers/content_security_policy.rb
Rails.application.config.content_security_policy do |policy|
  policy.default_src :self, :https
  policy.font_src    :self, :https, :data
  policy.img_src     :self, :https, :data
  policy.object_src  :none
  policy.script_src  :self, :https
  policy.style_src   :self, :https, :unsafe_inline
end
```

## SQL インジェクション対策

### ActiveRecordのプリペアドステートメント

ActiveRecordは自動的にプリペアドステートメントを使用します。

```ruby
# ✅ Good: 安全
User.where(email: params[:email])
User.where('email = ?', params[:email])
User.where('email = :email', email: params[:email])

# ❌ NG: SQLインジェクションの脆弱性
User.where("email = '#{params[:email]}'")
```

### where条件のプレースホルダー

必ずプレースホルダー（`?` または名前付き）を使用します。

```ruby
# ✅ Good: プレースホルダーを使用
def search_works(title)
  Work.where('title LIKE ?', "%#{sanitize_sql_like(title)}%")
end

# ❌ NG: 文字列補間
def search_works(title)
  Work.where("title LIKE '%#{title}%'")
end
```

## パスワード管理

### Deviseの使用

Rails版ではDeviseを使用してパスワードを管理します。

```ruby
# Deviseが自動的にbcryptでハッシュ化
user = User.create(
  email: 'user@example.com',
  password: 'secure_password',
  password_confirmation: 'secure_password'
)

# パスワードの検証
user.valid_password?('secure_password')  # true
```

### 平文パスワードの扱い

```ruby
# ❌ NG: 平文パスワードをログに出力
Rails.logger.info "User password: #{params[:password]}"

# ✅ Good: パスワードはログに出力しない
Rails.logger.info "User login attempt: #{params[:email]}"

# パラメータフィルタリング
# config/initializers/filter_parameter_logging.rb
Rails.application.config.filter_parameters += [:password, :password_confirmation]
```

## 認証・認可

### Deviseで認証

```ruby
# app/controllers/works_controller.rb
class WorksController < ApplicationController
  before_action :authenticate_user!, except: [:index, :show]

  def create
    # current_userはDeviseが提供
    @work = current_user.works.build(work_params)
    # ...
  end
end
```

### Punditで認可

```ruby
# app/controllers/works_controller.rb
class WorksController < ApplicationController
  before_action :set_work, only: [:edit, :update, :destroy]

  def update
    authorize @work  # Punditの認可チェック

    if @work.update(work_params)
      redirect_to @work, notice: '作品を更新しました'
    else
      render :edit
    end
  end

  private

  def set_work
    @work = Work.find(params[:id])
  end
end
```

## Strong Parameters

すべてのコントローラーでStrong Parametersを使用します。

```ruby
# ✅ Good: Strong Parameters
def work_params
  params.require(:work).permit(:title, :season_year, :season_name)
end

def create
  @work = Work.new(work_params)
  # ...
end

# ❌ NG: パラメータを直接使用
def create
  @work = Work.new(params[:work])  # Mass assignment vulnerability
end
```

## セッション管理

### セッションストア

Redis + ActiveRecordSessionStoreでセッションを管理します（Go版と共有）。

```ruby
# config/initializers/session_store.rb
Rails.application.config.session_store :active_record_store,
  key: '_annict_session',
  expire_after: 30.days
```

### セッションハイジャック対策

```ruby
# app/controllers/application_controller.rb
class ApplicationController < ActionController::Base
  before_action :configure_permitted_parameters, if: :devise_controller?
  before_action :reset_session_on_user_change

  private

  def reset_session_on_user_change
    return unless user_signed_in?
    return if session[:user_id] == current_user.id

    reset_session
    session[:user_id] = current_user.id
  end
end
```

## エラーメッセージ

### 詳細な情報を漏らさない

```ruby
# ❌ NG: 詳細なエラーメッセージをユーザーに表示
rescue ActiveRecord::RecordNotFound => e
  render json: { error: e.message, backtrace: e.backtrace }, status: :not_found
end

# ✅ Good: 一般的なエラーメッセージを表示
rescue ActiveRecord::RecordNotFound
  render json: { error: 'Not Found' }, status: :not_found
end

# サーバー側のログに詳細を記録
rescue => e
  Rails.logger.error "Error: #{e.message}"
  Rails.logger.error e.backtrace.join("\n")
  render json: { error: 'Internal Server Error' }, status: :internal_server_error
end
```

### Sentryでエラー追跡

```ruby
# config/initializers/sentry.rb
Sentry.init do |config|
  config.dsn = ENV['SENTRY_DSN']
  config.breadcrumbs_logger = [:active_support_logger, :http_logger]
  config.traces_sample_rate = 0.1
end

# app/controllers/application_controller.rb
class ApplicationController < ActionController::Base
  rescue_from StandardError, with: :handle_error

  private

  def handle_error(exception)
    Sentry.capture_exception(exception)
    render json: { error: 'Internal Server Error' }, status: :internal_server_error
  end
end
```

## セキュリティヘッダー

### Railsデフォルトのヘッダー

```ruby
# config/application.rb
module Annict
  class Application < Rails::Application
    # デフォルトで有効
    config.action_dispatch.default_headers = {
      'X-Frame-Options' => 'SAMEORIGIN',
      'X-Content-Type-Options' => 'nosniff',
      'X-XSS-Protection' => '1; mode=block'
    }
  end
end
```

### HSTSヘッダー（本番環境）

```ruby
# config/environments/production.rb
Rails.application.configure do
  config.force_ssl = true  # HSTSヘッダーを自動設定
end
```

## セキュリティチェックリスト

新機能を実装する際は、以下を必ず確認してください：

### フォーム送信

- [ ] `form_with` ヘルパーを使用しているか
- [ ] CSRFトークンが含まれているか

### ユーザー入力

- [ ] Strong Parametersを使用しているか
- [ ] バリデーションを実施しているか
- [ ] ホワイトリスト方式で許可しているか

### データベース

- [ ] ActiveRecordを使用しているか
- [ ] プレースホルダーを使用しているか
- [ ] 文字列補間を避けているか

### パスワード

- [ ] Deviseを使用しているか
- [ ] 平文パスワードをログに出力していないか
- [ ] パラメータフィルタリングを設定しているか

### 認証・認可

- [ ] `authenticate_user!` を設定しているか
- [ ] Punditで認可チェックを行っているか
- [ ] リソースの所有者チェックを行っているか

### エラー処理

- [ ] エラーメッセージは適切か
- [ ] 詳細な情報を漏らしていないか
- [ ] Sentryでエラーを追跡しているか

## ベストプラクティス

### 1. Bundler Auditで脆弱性をチェック

```sh
# 定期的に実行
bundle audit check --update
```

### 2. Brakemanで静的解析

```sh
# セキュリティ脆弱性をスキャン
brakeman -q
```

### 3. 環境変数でシークレットを管理

```ruby
# ❌ NG: ハードコード
API_KEY = 'abc123'

# ✅ Good: 環境変数
API_KEY = ENV['API_KEY']
```

### 4. 定期的にGemを更新

```sh
# 定期的に実行
bundle update
```

## トラブルシューティング

### CSRFトークンエラー

**症状**: "Can't verify CSRF token authenticity"

**原因**:
1. フォームにCSRFトークンが含まれていない
2. セッションが切れている
3. AJAX リクエストにトークンが含まれていない

**解決方法**:
```ruby
# フォーム: form_withを使用
<%= form_with model: @work do |f| %>
  <%= f.submit %>
<% end %>

# AJAX: meta タグからトークンを取得
// application.js
import Rails from '@rails/ujs'
Rails.start()

// これでAJAXリクエストに自動的にCSRFトークンが追加される
```

### Mass Assignment

**症状**: 想定外のパラメータが保存されてしまう

**原因**: Strong Parametersを使用していない

**解決方法**:
```ruby
# ✅ Good: Strong Parameters
def work_params
  params.require(:work).permit(:title, :season_year, :season_name)
end
```

## 参考資料

- [Rails Security Guide](https://guides.rubyonrails.org/security.html)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Brakeman](https://brakemanscanner.org/)
- [Bundler Audit](https://github.com/rubysec/bundler-audit)
