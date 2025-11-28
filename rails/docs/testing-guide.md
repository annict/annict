# テスト戦略ガイド（Rails版）

このドキュメントは、Rails版Annictのテスト戦略を説明します。

## 基本方針

- **テストファースト**: 実装前にテストを書くことを推奨
- **実データベースを使用**: 基本的にデータベースをモックせず、実際のPostgreSQLを使用
- **FactoryBot**: テストデータはFactoryBotで作成
- **カバレッジ**: SimpleCovでカバレッジを測定

## テストの種類

### モデルテスト（Model Specs）

- **配置**: `spec/models/`
- **責務**: モデルのバリデーション、メソッドの動作確認

#### 実装例

```ruby
# spec/models/work_spec.rb
require 'rails_helper'

RSpec.describe Work, type: :model do
  describe 'associations' do
    it { should belong_to(:created_by).class_name('User') }
    it { should have_many(:statuses) }
    it { should have_many(:records) }
  end

  describe 'validations' do
    it { should validate_presence_of(:title) }
    it { should validate_length_of(:title).is_at_most(255) }
  end

  describe 'scopes' do
    describe '.published' do
      it '公開されている作品のみ返す' do
        published_work = create(:work, published: true)
        unpublished_work = create(:work, published: false)

        expect(Work.published).to include(published_work)
        expect(Work.published).not_to include(unpublished_work)
      end
    end
  end

  describe '#popular' do
    it '人気順にソートされること' do
      work1 = create(:work, watchers_count: 100)
      work2 = create(:work, watchers_count: 200)
      work3 = create(:work, watchers_count: 150)

      expect(Work.popular).to eq([work2, work3, work1])
    end
  end

  describe '#season_name_i18n' do
    it 'シーズン名を国際化して返す' do
      work = build(:work, season_name: 'spring')
      expect(work.season_name_i18n).to eq('春')
    end
  end
end
```

### コントローラーテスト（Request Specs）

- **配置**: `spec/requests/`
- **責務**: HTTPリクエスト・レスポンスの確認、認証・認可のテスト

#### 実装例

```ruby
# spec/requests/works_spec.rb
require 'rails_helper'

RSpec.describe 'Works', type: :request do
  describe 'GET /works' do
    it 'ステータス200を返すこと' do
      get works_path
      expect(response).to have_http_status(:ok)
    end

    it '公開作品のみ表示すること' do
      published_work = create(:work, published: true, title: '公開作品')
      unpublished_work = create(:work, published: false, title: '非公開作品')

      get works_path

      expect(response.body).to include('公開作品')
      expect(response.body).not_to include('非公開作品')
    end
  end

  describe 'GET /works/:id' do
    let(:work) { create(:work) }

    it 'ステータス200を返すこと' do
      get work_path(work)
      expect(response).to have_http_status(:ok)
    end

    it '作品の詳細を表示すること' do
      get work_path(work)
      expect(response.body).to include(work.title)
    end
  end

  describe 'POST /works' do
    context 'ログインしていない場合' do
      it 'ログインページにリダイレクトされること' do
        post works_path, params: { work: attributes_for(:work) }
        expect(response).to redirect_to(new_user_session_path)
      end
    end

    context 'ログインしている場合' do
      let(:user) { create(:user) }
      before { sign_in user }

      context '有効なパラメータの場合' do
        it '作品を作成できること' do
          expect {
            post works_path, params: { work: attributes_for(:work) }
          }.to change(Work, :count).by(1)
        end

        it '作品詳細ページにリダイレクトされること' do
          post works_path, params: { work: attributes_for(:work) }
          expect(response).to redirect_to(work_path(Work.last))
        end

        it 'フラッシュメッセージが表示されること' do
          post works_path, params: { work: attributes_for(:work) }
          follow_redirect!
          expect(response.body).to include('作品を作成しました')
        end
      end

      context '無効なパラメータの場合' do
        it '作品を作成できないこと' do
          expect {
            post works_path, params: { work: { title: '' } }
          }.not_to change(Work, :count)
        end

        it 'エラーメッセージが表示されること' do
          post works_path, params: { work: { title: '' } }
          expect(response.body).to include('タイトルを入力してください')
        end
      end
    end
  end

  describe 'PATCH /works/:id' do
    let(:work) { create(:work) }

    context 'ログインしていない場合' do
      it 'ログインページにリダイレクトされること' do
        patch work_path(work), params: { work: { title: '新しいタイトル' } }
        expect(response).to redirect_to(new_user_session_path)
      end
    end

    context '作品の所有者の場合' do
      before { sign_in work.created_by }

      it '作品を更新できること' do
        patch work_path(work), params: { work: { title: '新しいタイトル' } }
        expect(work.reload.title).to eq('新しいタイトル')
      end
    end

    context '作品の所有者でない場合' do
      let(:other_user) { create(:user) }
      before { sign_in other_user }

      it '403エラーが返されること' do
        patch work_path(work), params: { work: { title: '新しいタイトル' } }
        expect(response).to have_http_status(:forbidden)
      end
    end
  end

  describe 'DELETE /works/:id' do
    let!(:work) { create(:work) }

    context '作品の所有者の場合' do
      before { sign_in work.created_by }

      it '作品を削除できること' do
        expect {
          delete work_path(work)
        }.to change(Work, :count).by(-1)
      end

      it '一覧ページにリダイレクトされること' do
        delete work_path(work)
        expect(response).to redirect_to(works_path)
      end
    end
  end
end
```

### システムテスト（System Specs / E2E）

- **配置**: `spec/system/`
- **責務**: ブラウザを使った実際のユーザー操作のテスト
- **ドライバー**: Capybara + Playwright

#### 実装例

```ruby
# spec/system/sign_in_spec.rb
require 'rails_helper'

RSpec.describe 'SignIn', type: :system do
  let(:user) { create(:user) }

  before do
    driven_by(:playwright)
  end

  describe 'ログイン' do
    it 'ユーザーがログインできること' do
      visit new_user_session_path

      fill_in 'Email', with: user.email
      fill_in 'Password', with: user.password
      click_button 'ログイン'

      expect(page).to have_content('ログインしました')
      expect(page).to have_content(user.username)
    end

    it '無効な認証情報でログインできないこと' do
      visit new_user_session_path

      fill_in 'Email', with: user.email
      fill_in 'Password', with: 'wrong_password'
      click_button 'ログイン'

      expect(page).to have_content('メールアドレスまたはパスワードが正しくありません')
    end
  end

  describe 'ログアウト' do
    before do
      sign_in user
      visit root_path
    end

    it 'ユーザーがログアウトできること' do
      click_link 'ログアウト'

      expect(page).to have_content('ログアウトしました')
      expect(page).to have_link('ログイン')
    end
  end
end
```

```ruby
# spec/system/works_spec.rb
require 'rails_helper'

RSpec.describe 'Works', type: :system do
  let(:user) { create(:user) }

  before do
    driven_by(:playwright)
    sign_in user
  end

  describe '作品の作成' do
    it '新しい作品を作成できること' do
      visit works_path
      click_link '新規作成'

      fill_in 'タイトル', with: 'テストアニメ'
      select '2024', from: 'シーズン年'
      select '春', from: 'シーズン'
      click_button '作成'

      expect(page).to have_content('作品を作成しました')
      expect(page).to have_content('テストアニメ')
    end

    it 'バリデーションエラーが表示されること' do
      visit new_work_path

      click_button '作成'

      expect(page).to have_content('タイトルを入力してください')
    end
  end

  describe '作品の編集' do
    let(:work) { create(:work, created_by: user) }

    it '作品を編集できること' do
      visit work_path(work)
      click_link '編集'

      fill_in 'タイトル', with: '新しいタイトル'
      click_button '更新'

      expect(page).to have_content('作品を更新しました')
      expect(page).to have_content('新しいタイトル')
    end
  end

  describe '作品の削除' do
    let!(:work) { create(:work, created_by: user) }

    it '作品を削除できること' do
      visit work_path(work)

      accept_confirm do
        click_link '削除'
      end

      expect(page).to have_content('作品を削除しました')
      expect(page).not_to have_content(work.title)
    end
  end
end
```

### GraphQL APIテスト

- **配置**: `spec/graphql/`
- **責務**: GraphQL APIのクエリ・ミューテーションのテスト

#### 実装例

```ruby
# spec/graphql/queries/works_query_spec.rb
require 'rails_helper'

RSpec.describe Queries::WorksQuery, type: :graphql do
  describe 'works' do
    let(:query) do
      <<~GQL
        query {
          works {
            nodes {
              id
              title
              watchersCount
            }
          }
        }
      GQL
    end

    it '作品一覧を取得できること' do
      works = create_list(:work, 3, published: true)
      result = execute_graphql(query)

      expect(result.dig('data', 'works', 'nodes').size).to eq(3)
    end

    it '非公開作品は含まれないこと' do
      create(:work, published: true, title: '公開作品')
      create(:work, published: false, title: '非公開作品')

      result = execute_graphql(query)
      nodes = result.dig('data', 'works', 'nodes')

      expect(nodes.size).to eq(1)
      expect(nodes.first['title']).to eq('公開作品')
    end
  end

  describe 'work' do
    let(:work) { create(:work) }
    let(:query) do
      <<~GQL
        query($id: ID!) {
          work(id: $id) {
            id
            title
            seasonYear
            seasonName
          }
        }
      GQL
    end

    it '作品の詳細を取得できること' do
      result = execute_graphql(query, variables: { id: work.id })
      work_data = result.dig('data', 'work')

      expect(work_data['id']).to eq(work.id.to_s)
      expect(work_data['title']).to eq(work.title)
    end

    it '存在しない作品の場合はnullを返すこと' do
      result = execute_graphql(query, variables: { id: 9999 })

      expect(result.dig('data', 'work')).to be_nil
      expect(result.dig('errors')).to be_present
    end
  end
end
```

## テストヘルパー

### FactoryBot

テストデータはFactoryBotで作成します。

```ruby
# spec/factories/works.rb
FactoryBot.define do
  factory :work do
    sequence(:title) { |n| "テストアニメ#{n}" }
    season_year { 2024 }
    season_name { 'spring' }
    published { true }
    association :created_by, factory: :user

    trait :unpublished do
      published { false }
    end

    trait :winter do
      season_name { 'winter' }
    end

    trait :with_image do
      after(:create) do |work|
        create(:work_image, work: work)
      end
    end
  end
end

# 使用例
work = create(:work)
unpublished_work = create(:work, :unpublished)
winter_work = create(:work, :winter, season_year: 2023)
work_with_image = create(:work, :with_image)
```

### データベースクリーンアップ

DatabaseCleanerを使用してテスト後のデータをクリーンアップします。

```ruby
# spec/rails_helper.rb
RSpec.configure do |config|
  config.use_transactional_fixtures = true

  config.before(:suite) do
    DatabaseCleaner.clean_with(:truncation)
  end

  config.before(:each) do
    DatabaseCleaner.strategy = :transaction
  end

  config.before(:each, type: :system) do
    DatabaseCleaner.strategy = :truncation
  end

  config.before(:each) do
    DatabaseCleaner.start
  end

  config.after(:each) do
    DatabaseCleaner.clean
  end
end
```

### Timecop

時間に依存するテストはTimecopを使用します。

```ruby
RSpec.describe Work do
  describe '#published_recently?' do
    it '最近公開された作品の場合はtrueを返す' do
      Timecop.freeze(Time.zone.local(2024, 1, 15)) do
        work = create(:work, published_at: 1.day.ago)
        expect(work.published_recently?).to be true
      end
    end

    it '古い作品の場合はfalseを返す' do
      Timecop.freeze(Time.zone.local(2024, 1, 15)) do
        work = create(:work, published_at: 1.month.ago)
        expect(work.published_recently?).to be false
      end
    end
  end
end
```

## ベストプラクティス

### 1. テストは読みやすく

```ruby
# ✅ Good: 明確なテストケース
it 'ログインユーザーは作品を作成できること' do
  sign_in user
  expect {
    post works_path, params: { work: attributes_for(:work) }
  }.to change(Work, :count).by(1)
end

# ❌ Bad: 何をテストしているかわかりにくい
it 'works' do
  sign_in user
  post works_path, params: { work: attributes_for(:work) }
  expect(Work.count).to eq(1)
end
```

### 2. 1つのテストで1つのことをテスト

```ruby
# ✅ Good: 1つのテストで1つのこと
it '作品を作成できること' do
  expect {
    create(:work)
  }.to change(Work, :count).by(1)
end

it '作品作成時に初期ステータスも作成されること' do
  work = create(:work)
  expect(work.statuses).to be_present
end

# ❌ Bad: 1つのテストで複数のことをテスト
it '作品を作成できること' do
  expect {
    work = create(:work)
    expect(work.statuses).to be_present
    expect(work.published).to be true
  }.to change(Work, :count).by(1)
end
```

### 3. let!とletを使い分ける

```ruby
# ✅ Good: 遅延評価を活用
let(:user) { create(:user) }  # テスト内で最初に使われたときに作成

# ✅ Good: 事前に必要なデータ
let!(:published_work) { create(:work, published: true) }  # すぐに作成

# ❌ Bad: 必要ないデータまで作成
before do
  @user1 = create(:user)
  @user2 = create(:user)
  @work1 = create(:work)
  @work2 = create(:work)
  # テストで使われない変数がある
end
```

### 4. コンテキストを活用

```ruby
# ✅ Good: コンテキストで条件を明確に
describe 'POST /works' do
  context 'ログインしていない場合' do
    it 'ログインページにリダイレクトされること' do
      # ...
    end
  end

  context 'ログインしている場合' do
    before { sign_in user }

    context '有効なパラメータの場合' do
      it '作品を作成できること' do
        # ...
      end
    end

    context '無効なパラメータの場合' do
      it 'エラーメッセージが表示されること' do
        # ...
      end
    end
  end
end
```

## トラブルシューティング

### テストが遅い

**原因**: 不要なデータを大量に作成している

**解決方法**:
- `let` を使って遅延評価
- `build` や `build_stubbed` を活用（DBに保存しない）
- 並列実行を検討（`parallel_tests` gem）

### ランダムに失敗するテスト

**原因**: テスト間でデータが共有されている

**解決方法**:
- DatabaseCleanerの設定を確認
- `let!` の使用を見直す
- Timecopの後始末を確認

### システムテストが失敗する

**原因**: JavaScriptの非同期処理を待っていない

**解決方法**:
```ruby
# ❌ Bad
click_button '送信'
expect(page).to have_content('成功')

# ✅ Good
click_button '送信'
expect(page).to have_content('成功', wait: 5)  # 最大5秒待つ
```

## 参考資料

- [RSpec公式ドキュメント](https://rspec.info/)
- [FactoryBot公式ドキュメント](https://github.com/thoughtbot/factory_bot)
- [Capybara公式ドキュメント](https://github.com/teamcapybara/capybara)
