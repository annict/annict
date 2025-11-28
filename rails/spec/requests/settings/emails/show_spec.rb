# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/email", type: :request do
  it "ログイン済みユーザーはページが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/email"

    expect(response.status).to eq(200)
    expect(response.body).to include("メールアドレス")
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    get "/settings/email"

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ユーザーの現在のメールアドレスが表示されること" do
    user = create(:registered_user, email: "test@example.com")
    login_as(user, scope: :user)

    get "/settings/email"

    expect(response.status).to eq(200)
    expect(response.body).to include("test@example.com")
  end
end
