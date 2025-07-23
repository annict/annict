# typed: false
# frozen_string_literal: true

RSpec.describe "GET /forum", type: :request do
  it "フォーラムポストの一覧が表示されること" do
    user = FactoryBot.create(:user)
    category1 = FactoryBot.create(:forum_category, :general)
    category2 = FactoryBot.create(:forum_category, :feedback)

    # 最新のコメント日時順に表示されるようにテストデータを作成
    post1 = FactoryBot.create(:forum_post, forum_category: category1, user:, last_commented_at: 3.days.ago, locale: "ja")
    post2 = FactoryBot.create(:forum_post, forum_category: category2, user:, last_commented_at: 1.day.ago, locale: "ja")
    post3 = FactoryBot.create(:forum_post, forum_category: category1, user:, last_commented_at: 2.days.ago, locale: "ja")

    get "/forum"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(post2.title)
    expect(response.body).to include(post3.title)
    expect(response.body).to include(post1.title)
  end

  it "削除されたユーザーのポストは表示されないこと" do
    user = FactoryBot.create(:user)
    deleted_user = FactoryBot.create(:user, deleted_at: Time.current)
    category = FactoryBot.create(:forum_category, :general)

    post1 = FactoryBot.create(:forum_post, forum_category: category, user:, locale: "ja")
    post2 = FactoryBot.create(:forum_post, forum_category: category, user: deleted_user, locale: "ja")

    get "/forum"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(post1.title)
    expect(response.body).not_to include(post2.title)
  end

  it "ページネーションが機能すること" do
    user = FactoryBot.create(:user)
    category = FactoryBot.create(:forum_category, :general)

    # デフォルトのページサイズ（25件）を超える数のポストを作成
    30.times do |i|
      FactoryBot.create(:forum_post,
        forum_category: category,
        user:,
        title: "ポスト #{i + 1}",
        last_commented_at: (30 - i).days.ago,
        locale: "ja")
    end

    get "/forum"

    expect(response).to have_http_status(:ok)
    # 最新のポストが表示されることを確認
    expect(response.body).to include("ポスト 30")
    # 26番目より古いポストは最初のページには表示されないことを確認
    expect(response.body).not_to include("ポスト 5")

    # 2ページ目を確認
    get "/forum?page=2"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("ポスト 5")
  end

  it "カテゴリ情報が含まれていること" do
    user = FactoryBot.create(:user)
    category = FactoryBot.create(:forum_category, :site_news)
    post = FactoryBot.create(:forum_post, forum_category: category, user:, locale: "ja")

    get "/forum"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(post.title)
    expect(response.body).to include(category.name)
  end

  it "ロケールによってリソースがフィルタリングされること" do
    user = FactoryBot.create(:user)
    category = FactoryBot.create(:forum_category, :general)

    FactoryBot.create(:forum_post,
      forum_category: category,
      user:,
      locale: "ja",
      title: "日本語のポスト")
    FactoryBot.create(:forum_post,
      forum_category: category,
      user:,
      locale: "en",
      title: "English Post")

    # デフォルト（日本語）でアクセス
    get "/forum"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("日本語のポスト")
    expect(response.body).not_to include("English Post")

    # 英語ドメインでアクセスする場合のテスト
    # localable_resourcesはドメインベースまたはユーザーの設定に基づいてフィルタリングされる
    # Accept-Languageヘッダーだけでは切り替わらないため、
    # ログインユーザーでロケール設定をテストする
    en_user = FactoryBot.create(:registered_user, locale: "en")
    login_as(en_user, scope: :user)

    get "/forum"

    expect(response).to have_http_status(:ok)
    # ログインユーザーは自分の投稿と許可されたロケールの投稿を見ることができる
    expect(response.body).to include("English Post")
  end
end
