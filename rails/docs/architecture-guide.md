# アーキテクチャガイド（Rails版）

このドキュメントは、Rails版Annictのアーキテクチャパターンを説明します。

## 概要

Rails版Annictは、標準のMVCアーキテクチャに加え、サービス層やコンポーネントを導入した構造を採用しています。

## ViewComponent

再利用可能なUIコンポーネントはViewComponentで実装します。

### 基本情報

- **ファイル配置**: `app/components/` ディレクトリ
- **命名規則**: `{ComponentName}Component` （例: `ButtonComponent`）
- **テンプレート**: Slimを使用
- **テスト**: RSpecでコンポーネントの動作をテスト

### 実装例

```ruby
# app/components/button_component.rb
# typed: true

class ButtonComponent < ViewComponent::Base
  extend T::Sig

  sig { params(text: String, url: String, options: T::Hash[Symbol, T.untyped]).void }
  def initialize(text:, url:, **options)
    @text = text
    @url = url
    @options = options
  end

  attr_reader :text, :url, :options
end
```

```slim
/ app/components/button_component.html.slim
= link_to @text, @url, class: "btn btn-primary", **@options
```

### ビューでの使用

```slim
/ app/views/works/show.html.slim
= render ButtonComponent.new(text: "編集", url: edit_work_path(@work))
```

### テスト

```ruby
# spec/components/button_component_spec.rb
require 'rails_helper'

RSpec.describe ButtonComponent, type: :component do
  it 'ボタンをレンダリングする' do
    component = ButtonComponent.new(text: "クリック", url: "/test")
    render_inline(component)

    expect(page).to have_link("クリック", href: "/test")
    expect(page).to have_css(".btn.btn-primary")
  end

  it 'オプションを渡せる' do
    component = ButtonComponent.new(
      text: "クリック",
      url: "/test",
      class: "custom-class"
    )
    render_inline(component)

    expect(page).to have_css(".custom-class")
  end
end
```

### ベストプラクティス

#### 1. 小さく保つ

```ruby
# ✅ Good: 単一責任
class ButtonComponent < ViewComponent::Base
  # ボタンのレンダリングのみ
end

# ❌ Bad: 複雑すぎる
class FormComponent < ViewComponent::Base
  # フォーム全体をレンダリング（大きすぎる）
end
```

#### 2. 再利用性を重視

```ruby
# ✅ Good: 汎用的なコンポーネント
class CardComponent < ViewComponent::Base
  def initialize(title:, content:, **options)
    @title = title
    @content = content
    @options = options
  end
end

# どこでも使える
render CardComponent.new(title: "作品名", content: "説明")
render CardComponent.new(title: "ユーザー名", content: "プロフィール")
```

#### 3. スロットを活用

```ruby
# app/components/card_component.rb
class CardComponent < ViewComponent::Base
  renders_one :header
  renders_one :body
  renders_one :footer
end
```

```slim
/ app/views/works/show.html.slim
= render CardComponent.new do |c|
  - c.with_header do
    h2= @work.title
  - c.with_body do
    p= @work.description
  - c.with_footer do
    = link_to "詳細", work_path(@work)
```

## サービスオブジェクト

複雑なビジネスロジックはサービスオブジェクトに抽出します。

### 基本情報

- **ファイル配置**: `app/services/` ディレクトリ
- **命名規則**: `{Action}{Entity}Service` （例: `CreateWorkService`）
- **メソッド**: `call` メソッドを実装
- **責務**: トランザクション管理、複数モデルを跨ぐ処理

### 実装例

```ruby
# app/services/create_work_service.rb
# typed: true

class CreateWorkService
  extend T::Sig

  sig { params(params: T::Hash[Symbol, T.untyped], user: User).returns(Work) }
  def self.call(params:, user:)
    ActiveRecord::Base.transaction do
      work = Work.create!(params.merge(created_by: user))

      # 関連レコードの作成
      work.create_initial_status!

      # イベントの発行
      WorkCreatedEvent.publish(work)

      work
    end
  end
end
```

### コントローラーでの使用

```ruby
# app/controllers/works_controller.rb
class WorksController < ApplicationController
  def create
    @work = CreateWorkService.call(
      params: work_params,
      user: current_user
    )

    redirect_to @work, notice: '作品を作成しました'
  rescue ActiveRecord::RecordInvalid => e
    @work = e.record
    render :new
  end

  private

  def work_params
    params.require(:work).permit(:title, :season_year, :season_name)
  end
end
```

### テスト

```ruby
# spec/services/create_work_service_spec.rb
require 'rails_helper'

RSpec.describe CreateWorkService do
  describe '.call' do
    let(:user) { create(:user) }
    let(:params) { { title: 'テストアニメ', season_year: 2024, season_name: 'spring' } }

    it '作品を作成する' do
      expect {
        CreateWorkService.call(params: params, user: user)
      }.to change(Work, :count).by(1)
    end

    it '初期ステータスを作成する' do
      work = CreateWorkService.call(params: params, user: user)
      expect(work.statuses).to be_present
    end

    it 'トランザクション内で実行される' do
      allow(Work).to receive(:create!).and_raise(ActiveRecord::RecordInvalid)

      expect {
        CreateWorkService.call(params: params, user: user)
      }.to raise_error(ActiveRecord::RecordInvalid)
        .and change(Work, :count).by(0)
    end
  end
end
```

### ベストプラクティス

#### 1. 単一責任の原則

```ruby
# ✅ Good: 1つの責務
class CreateWorkService
  def self.call(params:, user:)
    # 作品の作成のみ
  end
end

# ❌ Bad: 複数の責務
class WorkService
  def self.create(params:, user:)
  end

  def self.update(work:, params:)
  end

  def self.delete(work:)
  end
end
```

#### 2. クラスメソッドを使用

```ruby
# ✅ Good: クラスメソッド
CreateWorkService.call(params: params, user: user)

# ❌ Bad: インスタンスメソッド
service = CreateWorkService.new(params: params, user: user)
service.call
```

#### 3. Result オブジェクトを返す

```ruby
# ✅ Good: 成功・失敗を明示
class CreateWorkService
  Result = Struct.new(:success?, :work, :errors)

  def self.call(params:, user:)
    work = Work.create(params.merge(created_by: user))

    if work.persisted?
      Result.new(true, work, nil)
    else
      Result.new(false, work, work.errors)
    end
  end
end

# コントローラー
result = CreateWorkService.call(params: params, user: user)
if result.success?
  redirect_to result.work
else
  @work = result.work
  render :new
end
```

## Pundit（認可）

認可ロジックはPunditポリシーで管理します。

### 基本情報

- **ファイル配置**: `app/policies/` ディレクトリ
- **命名規則**: `{Model}Policy` （例: `WorkPolicy`）
- **メソッド**: `index?`, `show?`, `create?`, `update?`, `destroy?` など
- **コントローラー**: `authorize` メソッドで認可チェック

### 実装例

```ruby
# app/policies/work_policy.rb
# typed: true

class WorkPolicy < ApplicationPolicy
  extend T::Sig

  sig { returns(T::Boolean) }
  def index?
    true  # 誰でも一覧を見れる
  end

  sig { returns(T::Boolean) }
  def show?
    true  # 誰でも詳細を見れる
  end

  sig { returns(T::Boolean) }
  def create?
    user.present?  # ログインユーザーのみ作成可能
  end

  sig { returns(T::Boolean) }
  def update?
    user.present? && (record.created_by == user || user.admin?)
  end

  sig { returns(T::Boolean) }
  def destroy?
    user.present? && (record.created_by == user || user.admin?)
  end

  class Scope < Scope
    sig { returns(ActiveRecord::Relation) }
    def resolve
      if user&.admin?
        scope.all
      else
        scope.where(published: true)
      end
    end
  end
end
```

### コントローラーでの使用

```ruby
# app/controllers/works_controller.rb
class WorksController < ApplicationController
  before_action :set_work, only: [:show, :edit, :update, :destroy]

  def index
    @works = policy_scope(Work)
  end

  def show
    authorize @work
  end

  def new
    @work = Work.new
    authorize @work
  end

  def create
    @work = Work.new(work_params)
    authorize @work

    if @work.save
      redirect_to @work, notice: '作品を作成しました'
    else
      render :new
    end
  end

  def update
    authorize @work

    if @work.update(work_params)
      redirect_to @work, notice: '作品を更新しました'
    else
      render :edit
    end
  end

  def destroy
    authorize @work
    @work.destroy
    redirect_to works_path, notice: '作品を削除しました'
  end

  private

  def set_work
    @work = Work.find(params[:id])
  end

  def work_params
    params.require(:work).permit(:title, :season_year, :season_name)
  end
end
```

### ビューでの使用

```slim
/ app/views/works/show.html.slim
h1= @work.title

- if policy(@work).update?
  = link_to '編集', edit_work_path(@work), class: 'btn btn-primary'

- if policy(@work).destroy?
  = link_to '削除', work_path(@work), method: :delete, data: { confirm: '本当に削除しますか？' }, class: 'btn btn-danger'
```

### テスト

```ruby
# spec/policies/work_policy_spec.rb
require 'rails_helper'

RSpec.describe WorkPolicy do
  subject { described_class.new(user, work) }

  let(:work) { create(:work, created_by: owner) }
  let(:owner) { create(:user) }

  context 'ゲストユーザーの場合' do
    let(:user) { nil }

    it { should permit_action(:index) }
    it { should permit_action(:show) }
    it { should forbid_action(:create) }
    it { should forbid_action(:update) }
    it { should forbid_action(:destroy) }
  end

  context 'ログインユーザーの場合' do
    let(:user) { create(:user) }

    it { should permit_action(:index) }
    it { should permit_action(:show) }
    it { should permit_action(:create) }
    it { should forbid_action(:update) }
    it { should forbid_action(:destroy) }
  end

  context '作品の所有者の場合' do
    let(:user) { owner }

    it { should permit_action(:index) }
    it { should permit_action(:show) }
    it { should permit_action(:create) }
    it { should permit_action(:update) }
    it { should permit_action(:destroy) }
  end

  context '管理者の場合' do
    let(:user) { create(:user, admin: true) }

    it { should permit_action(:index) }
    it { should permit_action(:show) }
    it { should permit_action(:create) }
    it { should permit_action(:update) }
    it { should permit_action(:destroy) }
  end
end
```

### ベストプラクティス

#### 1. ポリシーメソッドをシンプルに

```ruby
# ✅ Good: シンプルなロジック
def update?
  user.present? && (record.created_by == user || user.admin?)
end

# ❌ Bad: 複雑すぎる
def update?
  return false unless user.present?
  return true if user.admin?
  return true if record.created_by == user
  return true if user.moderator? && record.created_at > 1.day.ago
  false
end
```

#### 2. Scopeを活用

```ruby
# ✅ Good: Scopeでフィルタリング
class WorkPolicy < ApplicationPolicy
  class Scope < Scope
    def resolve
      if user&.admin?
        scope.all
      else
        scope.where(published: true)
      end
    end
  end
end

# コントローラー
@works = policy_scope(Work)
```

#### 3. 認可エラーをハンドリング

```ruby
# app/controllers/application_controller.rb
class ApplicationController < ActionController::Base
  include Pundit::Authorization

  rescue_from Pundit::NotAuthorizedError, with: :user_not_authorized

  private

  def user_not_authorized
    flash[:alert] = 'この操作を実行する権限がありません'
    redirect_to(request.referrer || root_path)
  end
end
```

## まとめ

Rails版Annictでは、以下のパターンを活用してコードを整理しています：

- **ViewComponent**: 再利用可能なUIコンポーネント
- **サービスオブジェクト**: 複雑なビジネスロジックとトランザクション管理
- **Pundit**: 認可ロジックの管理

これらを適切に使い分けることで、MVCアーキテクチャをより保守しやすく拡張しやすいものにしています。
