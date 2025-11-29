# typed: false
# frozen_string_literal: true

RSpec.describe "GET /oauth/authorize", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    oauth_application = FactoryBot.create(:oauth_application)

    get "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code"
    }

    expect(response.status).to eq(302)
    expect(response.location).to start_with("http://api.annict.test:3000/sign_in")
  end

  it "ログインしているとき、認可パラメータが正しい場合、認可画面が表示されること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    get "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code"
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("認証")
  end

  it "ログインしているとき、client_idが無効な場合、エラーページが表示されること" do
    user = FactoryBot.create(:registered_user)

    login_as(user, scope: :user)

    get "/oauth/authorize", params: {
      client_id: "invalid_client_id",
      redirect_uri: "https://example.com",
      response_type: "code"
    }

    expect(response.status).to eq(401)
  end

  it "ログインしているとき、redirect_uriが無効な場合、エラーページが表示されること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    get "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: "https://invalid.example.com",
      response_type: "code"
    }

    expect(response.status).to eq(400)
  end

  it "ログインしているとき、response_typeが無効な場合、エラーページが表示されること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    get "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "invalid_type"
    }

    expect(response.status).to eq(400)
  end

  it "ログインしているとき、scopeパラメータが指定されている場合、認可画面が表示されること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    get "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code",
      scope: "read write"
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("認証")
  end

  it "ログインしているとき、stateパラメータが指定されている場合、認可画面が表示されること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    get "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code",
      state: "random_state_value"
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("認証")
  end
end
