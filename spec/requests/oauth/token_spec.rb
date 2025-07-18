# typed: false
# frozen_string_literal: true

RSpec.describe "POST /oauth/token", type: :request do
  it "authorization_codeグラントタイプで有効なコードを送信したとき、アクセストークンを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application, redirect_uri: "http://example.com/callback")
    access_grant = FactoryBot.create(:oauth_access_grant,
      resource_owner_id: user.id,
      application: oauth_application,
      redirect_uri: oauth_application.redirect_uri,
      expires_in: 600,
      scopes: "read")

    post "/oauth/token", params: {
      grant_type: "authorization_code",
      code: access_grant.token,
      redirect_uri: oauth_application.redirect_uri,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["access_token"]).to be_present
    expect(json["token_type"]).to eq("Bearer")
    expect(json["scope"]).to eq("read")
    expect(json["created_at"]).to be_present
  end

  it "authorization_codeグラントタイプで無効なコードを送信したとき、エラーを返すこと" do
    oauth_application = FactoryBot.create(:oauth_application, redirect_uri: "http://example.com/callback")

    post "/oauth/token", params: {
      grant_type: "authorization_code",
      code: "invalid_code",
      redirect_uri: oauth_application.redirect_uri,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:bad_request)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_grant")
    expect(json["error_description"]).to be_present
  end

  it "authorization_codeグラントタイプで期限切れのコードを送信したとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application, redirect_uri: "http://example.com/callback")
    access_grant = FactoryBot.create(:oauth_access_grant,
      resource_owner_id: user.id,
      application: oauth_application,
      redirect_uri: oauth_application.redirect_uri,
      expires_in: 600,
      scopes: "read",
      created_at: 11.minutes.ago)

    post "/oauth/token", params: {
      grant_type: "authorization_code",
      code: access_grant.token,
      redirect_uri: oauth_application.redirect_uri,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:bad_request)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_grant")
  end

  it "authorization_codeグラントタイプでリダイレクトURIが一致しないとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application, redirect_uri: "http://example.com/callback")
    access_grant = FactoryBot.create(:oauth_access_grant,
      resource_owner_id: user.id,
      application: oauth_application,
      redirect_uri: oauth_application.redirect_uri,
      expires_in: 600,
      scopes: "read")

    post "/oauth/token", params: {
      grant_type: "authorization_code",
      code: access_grant.token,
      redirect_uri: "http://wrong.com/callback",
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:bad_request)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_grant")
  end

  it "authorization_codeグラントタイプで無効なclient_idを送信したとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application, redirect_uri: "http://example.com/callback")
    access_grant = FactoryBot.create(:oauth_access_grant,
      resource_owner_id: user.id,
      application: oauth_application,
      redirect_uri: oauth_application.redirect_uri,
      expires_in: 600,
      scopes: "read")

    post "/oauth/token", params: {
      grant_type: "authorization_code",
      code: access_grant.token,
      redirect_uri: oauth_application.redirect_uri,
      client_id: "invalid_client_id",
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:unauthorized)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_client")
  end

  it "authorization_codeグラントタイプで無効なclient_secretを送信したとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application, redirect_uri: "http://example.com/callback")
    access_grant = FactoryBot.create(:oauth_access_grant,
      resource_owner_id: user.id,
      application: oauth_application,
      redirect_uri: oauth_application.redirect_uri,
      expires_in: 600,
      scopes: "read")

    post "/oauth/token", params: {
      grant_type: "authorization_code",
      code: access_grant.token,
      redirect_uri: oauth_application.redirect_uri,
      client_id: oauth_application.uid,
      client_secret: "invalid_secret"
    }

    expect(response).to have_http_status(:unauthorized)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_client")
  end

  it "サポートされていないgrant_typeを送信したとき、エラーを返すこと" do
    oauth_application = FactoryBot.create(:oauth_application)

    post "/oauth/token", params: {
      grant_type: "password",
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret,
      username: "test@example.com",
      password: "password"
    }

    expect(response).to have_http_status(:bad_request)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("unsupported_grant_type")
  end

  it "grant_typeパラメータが含まれていないとき、エラーを返すこと" do
    oauth_application = FactoryBot.create(:oauth_application)

    post "/oauth/token", params: {
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:bad_request)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_request")
  end

  it "複数のスコープが許可されたアクセスグラントで、正しいスコープを持つアクセストークンを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application, redirect_uri: "http://example.com/callback", scopes: "read write")
    access_grant = FactoryBot.create(:oauth_access_grant,
      resource_owner_id: user.id,
      application: oauth_application,
      redirect_uri: oauth_application.redirect_uri,
      expires_in: 600,
      scopes: "read write")

    post "/oauth/token", params: {
      grant_type: "authorization_code",
      code: access_grant.token,
      redirect_uri: oauth_application.redirect_uri,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["access_token"]).to be_present
    expect(json["scope"]).to eq("read write")
  end

  it "同じauthorization codeで2回目のリクエストをしたとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application, redirect_uri: "http://example.com/callback")
    access_grant = FactoryBot.create(:oauth_access_grant,
      resource_owner_id: user.id,
      application: oauth_application,
      redirect_uri: oauth_application.redirect_uri,
      expires_in: 600,
      scopes: "read")

    # 1回目のリクエスト
    post "/oauth/token", params: {
      grant_type: "authorization_code",
      code: access_grant.token,
      redirect_uri: oauth_application.redirect_uri,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)

    # 2回目のリクエスト（同じコードを使用）
    post "/oauth/token", params: {
      grant_type: "authorization_code",
      code: access_grant.token,
      redirect_uri: oauth_application.redirect_uri,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:bad_request)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_grant")
  end
end
