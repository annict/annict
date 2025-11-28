# typed: false
# frozen_string_literal: true

RSpec.describe "GET /forum/posts/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/forum/posts/new"

    expect(response).to redirect_to(new_user_session_path)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "ログインしているとき、新規投稿ページが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/forum/posts/new"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("投稿する")
  end

  it "ログインしているとき、カテゴリーパラメータが指定されている場合、そのカテゴリーが選択されること" do
    user = create(:registered_user)
    category = create(:forum_category, :general)
    login_as(user, scope: :user)

    get "/forum/posts/new", params: {category: category.slug}

    expect(response).to have_http_status(:ok)
    # カテゴリーが選択されていることを確認
    expect(response.body).to include("selected")
    expect(response.body).to include(category.id.to_s)
  end

  it "ログインしているとき、存在しないカテゴリーパラメータが指定されている場合、カテゴリーが設定されないこと" do
    user = create(:registered_user)
    create(:forum_category, :general)
    login_as(user, scope: :user)

    get "/forum/posts/new", params: {category: "non-existent-category"}

    expect(response).to have_http_status(:ok)
    # カテゴリーが選択されていないことを確認（デフォルトの選択状態を確認）
    expect(response.body).not_to include("selected")
  end

  it "ログインしているとき、site_newsカテゴリーが指定されても、一般ユーザーのフォームには表示されないこと" do
    user = create(:registered_user)
    site_news_category = create(:forum_category, :site_news)
    general_category = create(:forum_category, :general)
    login_as(user, scope: :user)

    get "/forum/posts/new", params: {category: site_news_category.slug}

    expect(response).to have_http_status(:ok)
    # site_newsカテゴリーはフォームの選択肢に含まれない
    expect(response.body).not_to include(site_news_category.local_name)
    # generalカテゴリーは選択肢に含まれる
    expect(response.body).to include(general_category.local_name)
  end

  it "ログインしているとき、管理者がsite_newsカテゴリーを指定した場合、正しく設定されること" do
    admin_user = create(:registered_user, :with_admin_role)
    site_news_category = create(:forum_category, :site_news)
    login_as(admin_user, scope: :user)

    get "/forum/posts/new", params: {category: site_news_category.slug}

    expect(response).to have_http_status(:ok)
    # site_newsカテゴリーが選択されていることを確認
    expect(response.body).to include("selected")
    expect(response.body).to include(site_news_category.id.to_s)
  end

  it "ログインしているとき、適切なメタタグが設定されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/forum/posts/new"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("<title>")
  end

  it "ログインしているとき、フォームが正しくレンダリングされること" do
    user = create(:registered_user)
    create(:forum_category, :general)
    login_as(user, scope: :user)

    get "/forum/posts/new"

    expect(response).to have_http_status(:ok)
    # フォームの存在を確認
    expect(response.body).to include("forum_post[title]")
    expect(response.body).to include("forum_post[body]")
    expect(response.body).to include("forum_post[forum_category_id]")
  end
end
