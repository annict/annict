# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/apps", type: :request do
  it "ログインしているとき、接続済みアプリとトークンが表示されること" do
    user = create(:registered_user)
    oauth_application = create(:oauth_application)
    create(:oauth_access_token, application: oauth_application, owner: user)
    login_as(user, scope: :user)

    get "/settings/apps"

    expect(response.status).to eq(200)
    expect(response.body).to include(oauth_application.name)
  end

  it "ログインしているとき、個人トークンが表示されること" do
    user = create(:registered_user)
    personal_token = create(:oauth_access_token, application: nil, owner: user, description: "Test Token")
    login_as(user, scope: :user)

    get "/settings/apps"

    expect(response.status).to eq(200)
    expect(response.body).to include("個人用アクセストークン")
    expect(response.body).to include(personal_token.description)
  end

  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/settings/apps"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end
end
