# typed: false
# frozen_string_literal: true

RSpec.describe "GET /oauth/token/info", type: :request do
  it "有効なアクセストークンが送信されたとき、トークン情報を返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read write")

    get "/oauth/token/info", headers: {
      "Authorization" => "Bearer #{access_token.token}"
    }

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["resource_owner_id"]).to eq(user.id)
    expect(json["scope"]).to eq(["read", "write"])
    expect(json["application"]["uid"]).to eq(oauth_application.uid)
    expect(json["expires_in"]).to be_nil
    expect(json["created_at"]).to be_present
  end

  it "無効なアクセストークンが送信されたとき、エラーを返すこと" do
    get "/oauth/token/info", headers: {
      "Authorization" => "Bearer invalid_token"
    }

    expect(response).to have_http_status(:unauthorized)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_token")
  end

  it "revoke済みのアクセストークンが送信されたとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read write",
      revoked_at: 1.hour.ago)

    get "/oauth/token/info", headers: {
      "Authorization" => "Bearer #{access_token.token}"
    }

    expect(response).to have_http_status(:unauthorized)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_token")
  end

  it "期限が設定されているアクセストークンが送信されたとき、期限情報を返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read write",
      expires_in: 3600)

    get "/oauth/token/info", headers: {
      "Authorization" => "Bearer #{access_token.token}"
    }

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["resource_owner_id"]).to eq(user.id)
    expect(json["scope"]).to eq(["read", "write"])
    expect(json["expires_in"]).to be_present
    expect(json["expires_in"]).to be <= 3600
  end

  it "期限切れのアクセストークンが送信されたとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read write",
      expires_in: 3600,
      created_at: 2.hours.ago)

    get "/oauth/token/info", headers: {
      "Authorization" => "Bearer #{access_token.token}"
    }

    expect(response).to have_http_status(:unauthorized)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_token")
  end

  it "アクセストークンが送信されていないとき、エラーを返すこと" do
    get "/oauth/token/info"

    expect(response).to have_http_status(:unauthorized)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_token")
  end

  it "限定されたスコープを持つアクセストークンが送信されたとき、そのスコープ情報を返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read")

    get "/oauth/token/info", headers: {
      "Authorization" => "Bearer #{access_token.token}"
    }

    expect(response).to have_http_status(:ok)
    json = JSON.parse(response.body)
    expect(json["resource_owner_id"]).to eq(user.id)
    expect(json["scope"]).to eq(["read"])
    expect(json["application"]["uid"]).to eq(oauth_application.uid)
  end

  it "Bearer トークンの形式でないAuthorizationヘッダーが送信されたとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      application: oauth_application,
      resource_owner_id: user.id,
      scopes: "read write")

    get "/oauth/token/info", headers: {
      "Authorization" => "Basic #{access_token.token}"
    }

    expect(response).to have_http_status(:unauthorized)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("invalid_token")
  end
end
