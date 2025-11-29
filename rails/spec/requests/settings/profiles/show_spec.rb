# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/profile", type: :request do
  it "認証されたユーザーがアクセスした場合、プロフィール設定ページが正常に表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/profile"

    expect(response.status).to eq(200)
    expect(response.body).to include("自己紹介")
  end

  it "認証されていないユーザーがアクセスした場合、ログインページにリダイレクトされること" do
    get "/settings/profile"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end
end
