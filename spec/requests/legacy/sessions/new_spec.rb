# typed: false
# frozen_string_literal: true

RSpec.describe "GET /legacy/sign_in", type: :request do
  it "ログインしていないとき、パスワードでログインページが表示されること" do
    get "/legacy/sign_in"

    expect(response.status).to eq(200)
    expect(response.body).to include("おかえりなさい！")
  end

  it "ログインしているとき、トップページにリダイレクトされること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/legacy/sign_in"

    expect(response.status).to eq(302)
    expect(response).to redirect_to("/")
  end

  it "backパラメータがあるとき、リダイレクト先として保存されること" do
    back_url = "/some/path"

    get "/legacy/sign_in", params: {back: back_url}

    expect(response.status).to eq(200)
    expect(response.body).to include("おかえりなさい！")
  end
end
