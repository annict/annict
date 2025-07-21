# typed: false
# frozen_string_literal: true

RSpec.describe "GET /forum/posts/:post_id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:)

    get forum_edit_post_path(forum_post)

    expect(response).to redirect_to(new_user_session_path)
  end

  it "投稿の作成者がログインしているとき、編集ページが表示されること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:, title: "編集するタイトル", body: "編集する本文")

    login_as(user, scope: :user)
    get forum_edit_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("編集するタイトル")
    expect(response.body).to include("編集する本文")
  end

  it "他のユーザーの投稿を編集しようとしたとき、403エラーになること" do
    post_owner = create(:registered_user)
    other_user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user: post_owner, forum_category:)

    login_as(other_user, scope: :user)

    expect {
      get forum_edit_post_path(forum_post)
    }.to raise_error(Pundit::NotAuthorizedError)
  end

  it "存在しない投稿を編集しようとしたとき、404エラーになること" do
    user = create(:registered_user)

    login_as(user, scope: :user)

    expect {
      get "/forum/posts/999999999/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "投稿者が削除されている投稿を編集しようとしたとき、403エラーになること" do
    deleted_user = create(:registered_user, deleted_at: Time.current)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user: deleted_user, forum_category:)

    # 削除されたユーザーはログインできないため、他のユーザーでテスト
    other_user = create(:registered_user)
    login_as(other_user, scope: :user)

    expect {
      get forum_edit_post_path(forum_post)
    }.to raise_error(Pundit::NotAuthorizedError)
  end

  it "日本語ロケールの投稿が編集ページで正しく表示されること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:, title: "日本語のタイトル", body: "日本語の本文です。", locale: "ja")

    login_as(user, scope: :user)
    get forum_edit_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("日本語のタイトル")
    expect(response.body).to include("日本語の本文です。")
  end

  it "英語ロケールの投稿が編集ページで正しく表示されること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:, title: "English Title", body: "This is English body.", locale: "en")

    login_as(user, scope: :user)
    get forum_edit_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("English Title")
    expect(response.body).to include("This is English body.")
  end

  it "改行を含む本文が編集ページで正しく表示されること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    body_with_breaks = "1行目\n\n3行目\n\n\n6行目"
    forum_post = create(:forum_post, user:, forum_category:, title: "改行テスト", body: body_with_breaks)

    login_as(user, scope: :user)
    get forum_edit_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("改行テスト")
    # テキストエリアの中身を確認
    expect(response.body).to include(body_with_breaks)
  end

  it "異なるカテゴリーの投稿も編集できること" do
    user = create(:registered_user)
    feedback_category = create(:forum_category, :feedback)
    forum_post = create(:forum_post, user:, forum_category: feedback_category, title: "フィードバック投稿")

    login_as(user, scope: :user)
    get forum_edit_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("フィードバック投稿")
  end

  it "管理者がsite_newsカテゴリーの自分の投稿を編集できること" do
    admin_user = create(:registered_user, :with_admin_role)
    site_news_category = create(:forum_category, :site_news)
    forum_post = create(:forum_post, user: admin_user, forum_category: site_news_category, title: "サイトニュース")

    login_as(admin_user, scope: :user)
    get forum_edit_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("サイトニュース")
  end

  it "管理者でも他のユーザーの投稿は編集できないこと" do
    admin_user = create(:registered_user, :with_admin_role)
    regular_user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user: regular_user, forum_category:)

    login_as(admin_user, scope: :user)

    expect {
      get forum_edit_post_path(forum_post)
    }.to raise_error(Pundit::NotAuthorizedError)
  end

  it "非常に長いタイトルと本文の投稿も編集ページで表示できること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    long_title = "あ" * 100
    long_body = "い" * 10000
    forum_post = create(:forum_post, user:, forum_category:, title: long_title, body: long_body)

    login_as(user, scope: :user)
    get forum_edit_post_path(forum_post)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(long_title[0..50])
    expect(response.body).to include(long_body[0..50])
  end
end
