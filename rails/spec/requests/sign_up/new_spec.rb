# typed: false
# frozen_string_literal: true

RSpec.describe "GET /sign_up", type: :request do
  it "ログインしていないとき、サインアップページが正常に表示されること" do
    get "/sign_up"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("アカウント作成")
  end

  it "既にログインしているユーザーの場合、トップページにリダイレクトされること" do
    user = FactoryBot.create(:user)
    login_as(user, scope: :user)

    get "/sign_up"

    expect(response).to redirect_to(root_path)
  end

  it "レスポンスヘッダーにContent-Typeが正しく設定されること" do
    get "/sign_up"

    expect(response.content_type).to start_with("text/html")
  end

  it "main_simpleレイアウトが使用されること" do
    get "/sign_up"

    expect(response.body).not_to include("class=\"main-layout\"")
  end
end
