# typed: false
# frozen_string_literal: true

RSpec.describe "GET /forum/categories/:category_id", type: :request do
  it "カテゴリーが存在しないとき、404エラーを返すこと" do
    get "/forum/categories/non_existent_category"

    expect(response.status).to eq(404)
  end

  it "カテゴリーが存在するとき、カテゴリー情報と投稿一覧を表示すること" do
    forum_category = FactoryBot.create(:forum_category, :general)
    user = FactoryBot.create(:registered_user)
    forum_post1 = FactoryBot.create(:forum_post, forum_category:, user:, title: "最初の投稿", last_commented_at: 2.hours.ago)
    forum_post2 = FactoryBot.create(:forum_post, forum_category:, user:, title: "2番目の投稿", last_commented_at: 1.hour.ago)

    get "/forum/categories/#{forum_category.slug}"

    expect(response.status).to eq(200)
    expect(response.body).to include(forum_category.name)
    expect(response.body).to include(forum_post1.title)
    expect(response.body).to include(forum_post2.title)
  end

  it "投稿が新しい順に表示されること" do
    forum_category = FactoryBot.create(:forum_category, :general)
    user = FactoryBot.create(:registered_user)
    old_post = FactoryBot.create(:forum_post, forum_category:, user:, title: "古い投稿", last_commented_at: 2.days.ago)
    new_post = FactoryBot.create(:forum_post, forum_category:, user:, title: "新しい投稿", last_commented_at: 1.hour.ago)

    get "/forum/categories/#{forum_category.slug}"

    expect(response.status).to eq(200)
    # 新しい投稿が古い投稿より先に表示されることを確認
    expect(response.body.index(new_post.title)).to be < response.body.index(old_post.title)
  end

  it "削除されたユーザーの投稿は表示されないこと" do
    forum_category = FactoryBot.create(:forum_category, :general)
    active_user = FactoryBot.create(:registered_user)
    deleted_user = FactoryBot.create(:registered_user)

    active_post = FactoryBot.create(:forum_post, forum_category:, user: active_user, title: "アクティブユーザーの投稿")
    deleted_post = FactoryBot.create(:forum_post, forum_category:, user: deleted_user, title: "削除されたユーザーの投稿")

    # ユーザーを削除
    deleted_user.destroy_in_batches

    get "/forum/categories/#{forum_category.slug}"

    expect(response.status).to eq(200)
    expect(response.body).to include(active_post.title)
    expect(response.body).not_to include(deleted_post.title)
  end

  it "投稿がないとき、投稿がない旨を表示すること" do
    forum_category = FactoryBot.create(:forum_category, :general)

    get "/forum/categories/#{forum_category.slug}"

    expect(response.status).to eq(200)
    expect(response.body).to include(forum_category.name)
    expect(response.body).to include("投稿はありません")
  end

  it "site_newsカテゴリーのとき、新規投稿ボタンが表示されないこと" do
    forum_category = FactoryBot.create(:forum_category, :site_news)

    get "/forum/categories/#{forum_category.slug}"

    expect(response.status).to eq(200)
    expect(response.body).not_to include("新規投稿")
  end

  it "site_news以外のカテゴリーのとき、新規投稿ボタンが表示されること" do
    forum_category = FactoryBot.create(:forum_category, :general)

    get "/forum/categories/#{forum_category.slug}"

    expect(response.status).to eq(200)
    expect(response.body).to include("新規作成")
    expect(response.body).to include("/forum/posts/new?category=#{forum_category.slug}")
  end

  it "ページネーションが動作すること" do
    forum_category = FactoryBot.create(:forum_category, :general)
    user = FactoryBot.create(:registered_user)

    # デフォルトのページサイズ以上の投稿を作成
    30.times do |i|
      FactoryBot.create(:forum_post, forum_category:, user:, title: "投稿#{i + 1}", last_commented_at: i.hours.ago)
    end

    get "/forum/categories/#{forum_category.slug}"

    expect(response.status).to eq(200)
    # ページネーションボタンが表示されることを確認
    expect(response.body).to include("btn-group")
  end
end
