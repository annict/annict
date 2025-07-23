# typed: false
# frozen_string_literal: true

RSpec.describe "GET /oauth/authorize/native", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    oauth_application = FactoryBot.create(:oauth_application)

    get "/oauth/authorize/native", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code"
    }

    expect(response.status).to eq(302)
    expect(response.location).to start_with("http://api.annict.test:3000/sign_in")
  end

  it "ログインしているとき、認可パラメータが正しい場合、認可コードが表示されること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    # 事前に認可を許可しておく
    post "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code"
    }

    get "/oauth/authorize/native", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code"
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("認証コード")
  end

  it "ログインしているとき、client_idが無効な場合、認証コードページが表示されるが認証コードは空であること" do
    user = FactoryBot.create(:registered_user)

    login_as(user, scope: :user)

    get "/oauth/authorize/native", params: {
      client_id: "invalid_client_id",
      redirect_uri: "https://example.com",
      response_type: "code"
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("認証コード")
    expect(response.body).to match(/<code[^>]*>\s*<\/code>/)  # 空のcodeタグを確認
  end

  it "ログインしているとき、redirect_uriが無効な場合、認証コードページが表示されるが認証コードは空であること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    get "/oauth/authorize/native", params: {
      client_id: oauth_application.uid,
      redirect_uri: "https://invalid.example.com",
      response_type: "code"
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("認証コード")
    expect(response.body).to match(/<code[^>]*>\s*<\/code>/)  # 空のcodeタグを確認
  end

  it "ログインしているとき、response_typeが無効な場合、認証コードページが表示されるが認証コードは空であること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    get "/oauth/authorize/native", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "invalid_type"
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("認証コード")
    expect(response.body).to match(/<code[^>]*>\s*<\/code>/)  # 空のcodeタグを確認
  end

  it "ログインしているとき、scopeパラメータが指定されている場合、認証コードが表示されること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    # 事前に認可を許可しておく
    post "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code",
      scope: "read write"
    }

    get "/oauth/authorize/native", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code",
      scope: "read write"
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("認証コード")
  end

  it "ログインしているとき、stateパラメータが指定されている場合、認証コードが表示されること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    # 事前に認可を許可しておく
    post "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code",
      state: "random_state_value"
    }

    get "/oauth/authorize/native", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code",
      state: "random_state_value"
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("認証コード")
  end
end
