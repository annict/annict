# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/password", type: :request do
  it "ログイン済みユーザーはパスワード変更ページが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/password"

    expect(response.status).to eq(200)
    expect(response.body).to include("パスワード")
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    get "/settings/password"

    expect(response).to redirect_to(new_user_session_path)
  end
end
