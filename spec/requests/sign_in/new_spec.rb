# typed: false
# frozen_string_literal: true

RSpec.describe "GET /sign_in", type: :request do
  it "ログインしているとき、トップページにリダイレクトされること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/sign_in"

    expect(response.status).to eq(302)
    expect(response).to redirect_to("/")
  end

  it "ログインしていないとき、ログインページが表示されること" do
    get "/sign_in"

    expect(response.status).to eq(200)
    expect(response.body).to include("おかえりなさい！")
  end

  it "backパラメータが提供されたとき、正常にページが表示されること" do
    get "/sign_in", params: {back: "/profile"}

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("おかえりなさい！")
  end

  it "有効なclient_idパラメータが提供されたとき、正常にページが表示されること" do
    user = create(:registered_user)
    oauth_app = create(:oauth_application, owner: user)

    get "/sign_in", params: {client_id: oauth_app.uid}

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("おかえりなさい！")
  end

  it "無効なclient_idパラメータが提供されたとき、正常にページが表示されること" do
    get "/sign_in", params: {client_id: "invalid_uid"}

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("おかえりなさい！")
  end

  it "backパラメータとclient_idパラメータが両方提供されたとき、正しく処理されること" do
    user = create(:registered_user)
    oauth_app = create(:oauth_application, owner: user)

    get "/sign_in", params: {back: "/profile", client_id: oauth_app.uid}

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("おかえりなさい！")
  end
end
