# typed: false
# frozen_string_literal: true

RSpec.describe "POST /oauth/introspect", type: :request do
  it "有効なアクセストークンが送信されたとき、トークン情報を返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read write")

    post "/oauth/introspect", params: {
      token: access_token.token,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["active"]).to be(true)
    expect(json["scope"]).to eq("read write")
    expect(json["client_id"]).to eq(oauth_application.uid)
    expect(json["exp"]).to be_present
  end

  it "無効なアクセストークンが送信されたとき、非アクティブ状態を返すこと" do
    oauth_application = FactoryBot.create(:oauth_application)

    post "/oauth/introspect", params: {
      token: "invalid_token",
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["active"]).to be(false)
  end

  it "revoke済みのアクセストークンが送信されたとき、非アクティブ状態を返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read write",
      revoked_at: 1.hour.ago)

    post "/oauth/introspect", params: {
      token: access_token.token,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["active"]).to be(false)
  end

  it "期限切れのアクセストークンが送信されたとき、非アクティブ状態を返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read write",
      expires_in: 3600,
      created_at: 2.hours.ago)

    post "/oauth/introspect", params: {
      token: access_token.token,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["active"]).to be(false)
  end

  it "無効なclient_idが送信されたとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read write")

    post "/oauth/introspect", params: {
      token: access_token.token,
      client_id: "invalid_client_id",
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:unauthorized)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_client")
  end

  it "無効なclient_secretが送信されたとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read write")

    post "/oauth/introspect", params: {
      token: access_token.token,
      client_id: oauth_application.uid,
      client_secret: "invalid_secret"
    }

    expect(response).to have_http_status(:unauthorized)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_client")
  end

  it "tokenパラメータが含まれていないとき、非アクティブ状態を返すこと" do
    oauth_application = FactoryBot.create(:oauth_application)

    post "/oauth/introspect", params: {
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["active"]).to be(false)
  end

  it "限定されたスコープを持つアクセストークンが送信されたとき、そのスコープ情報を返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read")

    post "/oauth/introspect", params: {
      token: access_token.token,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["active"]).to be(true)
    expect(json["scope"]).to eq("read")
  end

  it "client認証なしでリクエストしたとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read write")

    post "/oauth/introspect", params: {
      token: access_token.token
    }

    expect(response).to have_http_status(:bad_request)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_request")
  end
end
