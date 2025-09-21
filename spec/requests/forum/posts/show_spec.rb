# typed: false
# frozen_string_literal: true

RSpec.describe "GET /forum/posts/:post_id", type: :request do
  it "投稿が存在し、ユーザーが削除されていないとき、投稿とコメントが表示されること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:, title: "テストタイトル", body: "テスト本文")
    create(:forum_comment, forum_post:, user:, body: "最初のコメント")
    create(:forum_comment, forum_post:, user:, body: "2番目のコメント")

    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("テストタイトル")
    expect(response.body).to include("テスト本文")
    expect(response.body).to include("最初のコメント")
    expect(response.body).to include("2番目のコメント")
  end

  it "投稿が存在するが、投稿者が削除されているとき、404エラーになること" do
    deleted_user = create(:registered_user, deleted_at: Time.current)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user: deleted_user, forum_category:, title: "削除されたユーザーの投稿")

    get forum_post_path(forum_post)

    expect(response.status).to eq(404)
  end

  it "投稿が存在しないとき、404エラーになること" do
    # 数値のIDで存在しないものを指定する
    get "/forum/posts/999999999"

    expect(response.status).to eq(404)
  end

  it "ログインしていないユーザーでも投稿を閲覧できること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:, title: "公開投稿", body: "誰でも見れる内容")

    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("公開投稿")
    expect(response.body).to include("誰でも見れる内容")
  end

  it "ログインしているユーザーが投稿を閲覧できること" do
    viewer = create(:registered_user)
    poster = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user: poster, forum_category:, title: "ログインユーザー向け投稿")

    login_as(viewer, scope: :user)
    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("ログインユーザー向け投稿")
  end

  it "コメントがない投稿でも正常に表示されること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:, title: "コメントなし投稿", body: "まだコメントがありません")

    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("コメントなし投稿")
    expect(response.body).to include("まだコメントがありません")
  end

  it "コメントが作成順に表示されること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:)

    # 順番を明確にするため、created_atを明示的に設定
    create(:forum_comment, forum_post:, user:, body: "3番目", created_at: 3.days.ago)
    create(:forum_comment, forum_post:, user:, body: "1番目", created_at: 5.days.ago)
    create(:forum_comment, forum_post:, user:, body: "2番目", created_at: 4.days.ago)

    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    # HTMLの中でコメントが表示される順番を確認
    body = response.body
    index_1 = body.index("1番目")
    index_2 = body.index("2番目")
    index_3 = body.index("3番目")

    expect(index_1).to be < index_2
    expect(index_2).to be < index_3
  end

  it "異なるカテゴリーの投稿も表示できること" do
    user = create(:registered_user)
    feedback_category = create(:forum_category, :feedback)
    forum_post = create(:forum_post, user:, forum_category: feedback_category, title: "フィードバックカテゴリーの投稿")

    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("フィードバックカテゴリーの投稿")
  end

  it "site_newsカテゴリーの投稿も表示できること" do
    admin_user = create(:registered_user, :with_admin_role)
    site_news_category = create(:forum_category, :site_news)
    forum_post = create(:forum_post, user: admin_user, forum_category: site_news_category, title: "サイトアップデート情報")

    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("サイトアップデート情報")
  end

  it "削除されたコメントがあっても投稿は表示されること" do
    user = create(:registered_user)
    commenter = create(:registered_user, deleted_at: Time.current)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:, title: "削除されたコメント付き投稿")
    create(:forum_comment, forum_post:, user: commenter, body: "削除されたユーザーのコメント")
    create(:forum_comment, forum_post:, user:, body: "有効なコメント")

    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("削除されたコメント付き投稿")
    expect(response.body).to include("有効なコメント")
  end

  it "日本語ロケールの投稿が正しく表示されること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:, title: "日本語タイトル", body: "これは日本語の投稿です。", locale: "ja")

    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("日本語タイトル")
    expect(response.body).to include("これは日本語の投稿です。")
  end

  it "英語ロケールの投稿が正しく表示されること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:, title: "English Title", body: "This is an English post.", locale: "en")

    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("English Title")
    expect(response.body).to include("This is an English post.")
  end

  it "改行を含む本文が正しく表示されること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    body_with_breaks = "これは1行目です。\n\nこれは3行目です。\n\n\nこれは6行目です。"
    forum_post = create(:forum_post, user:, forum_category:, title: "改行テスト", body: body_with_breaks)

    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("これは1行目です。")
    expect(response.body).to include("これは3行目です。")
    expect(response.body).to include("これは6行目です。")
  end

  it "非常に長いタイトルと本文の投稿も表示できること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    long_title = "あ" * 100 # 最大文字数
    long_body = "い" * 10000 # 最大文字数
    forum_post = create(:forum_post, user:, forum_category:, title: long_title, body: long_body)

    get forum_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(long_title[0..50]) # 少なくとも一部が表示されることを確認
    expect(response.body).to include(long_body[0..50]) # 少なくとも一部が表示されることを確認
  end
end
