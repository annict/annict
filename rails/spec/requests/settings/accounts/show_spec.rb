# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/account", type: :request do
  it "ログイン済みユーザーはページが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/account"

    expect(response.status).to eq(200)
    expect(response.body).to include("基本情報")
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    get "/settings/account"

    expect(response).to redirect_to(new_user_session_path)
  end
end
