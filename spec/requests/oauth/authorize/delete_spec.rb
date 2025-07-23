# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /oauth/authorize", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    oauth_application = FactoryBot.create(:oauth_application)

    delete "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code"
    }

    expect(response.status).to eq(302)
    expect(response.location).to start_with("http://api.annict.test:3000/sign_in")
  end

  it "ログインしているとき、認可パラメータが正しい場合、認可を拒否してリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    delete "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code"
    }

    expect(response.status).to eq(302)
    expect(response.location).to start_with(oauth_application.redirect_uri)
    expect(response.location).to include("error=access_denied")
  end

  it "ログインしているとき、client_idが無効な場合、リダイレクトされること" do
    user = FactoryBot.create(:registered_user)

    login_as(user, scope: :user)

    delete "/oauth/authorize", params: {
      client_id: "invalid_client_id",
      redirect_uri: "https://example.com",
      response_type: "code"
    }

    expect(response.status).to eq(302)
    expect(response.location).to include("error=")
  end

  it "ログインしているとき、redirect_uriが無効な場合、リダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    delete "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: "https://invalid.example.com",
      response_type: "code"
    }

    expect(response.status).to eq(302)
    expect(response.location).to include("error=")
  end

  it "ログインしているとき、response_typeが無効な場合、エラーページが表示されること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    delete "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "invalid_type"
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("Error")
  end

  it "ログインしているとき、scopeパラメータが指定されている場合、認可を拒否してリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    delete "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code",
      scope: "read write"
    }

    expect(response.status).to eq(302)
    expect(response.location).to start_with(oauth_application.redirect_uri)
    expect(response.location).to include("error=access_denied")
  end

  it "ログインしているとき、stateパラメータが指定されている場合、認可を拒否してリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application, owner: user)

    login_as(user, scope: :user)

    delete "/oauth/authorize", params: {
      client_id: oauth_application.uid,
      redirect_uri: oauth_application.redirect_uri,
      response_type: "code",
      state: "random_state_value"
    }

    expect(response.status).to eq(302)
    expect(response.location).to start_with(oauth_application.redirect_uri)
    expect(response.location).to include("error=access_denied")
    expect(response.location).to include("state=random_state_value")
  end
end
