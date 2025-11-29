# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/tokens/new", type: :request do
  it "ログインしているとき、新しいトークンの作成ページが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/tokens/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("oauth_access_token[description]")
    expect(response.body).to include("oauth_access_token[scopes]")
  end

  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/settings/tokens/new"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end
end
