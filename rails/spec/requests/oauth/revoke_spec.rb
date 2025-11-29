# typed: false
# frozen_string_literal: true

RSpec.describe "POST /oauth/revoke", type: :request do
  it "有効なアクセストークンを送信したとき、トークンを無効化すること" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      resource_owner_id: user.id,
      application: oauth_application,
      scopes: "read")

    post "/oauth/revoke", params: {
      token: access_token.token,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    expect(response.body).to eq("{}")

    # トークンが無効化されていることを確認
    access_token.reload
    expect(access_token.revoked_at).to be_present
  end

  it "有効なリフレッシュトークンを送信したとき、トークンを無効化すること" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      resource_owner_id: user.id,
      application: oauth_application,
      scopes: "read",
      refresh_token: "refresh_token_123")

    post "/oauth/revoke", params: {
      token: access_token.refresh_token,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    expect(response.body).to eq("{}")

    # トークンが無効化されていることを確認
    access_token.reload
    expect(access_token.revoked_at).to be_present
  end

  it "無効なトークンを送信したとき、成功すること" do
    oauth_application = FactoryBot.create(:oauth_application)

    post "/oauth/revoke", params: {
      token: "invalid_token",
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    expect(response.body).to eq("{}")
  end

  it "無効なclient_idを送信したとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      resource_owner_id: user.id,
      application: oauth_application,
      scopes: "read")

    post "/oauth/revoke", params: {
      token: access_token.token,
      client_id: "invalid_client_id",
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:forbidden)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("unauthorized_client")
  end

  it "無効なclient_secretを送信したとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      resource_owner_id: user.id,
      application: oauth_application,
      scopes: "read")

    post "/oauth/revoke", params: {
      token: access_token.token,
      client_id: oauth_application.uid,
      client_secret: "invalid_secret"
    }

    expect(response).to have_http_status(:forbidden)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("unauthorized_client")
  end

  it "tokenパラメータが含まれていないとき、成功すること" do
    oauth_application = FactoryBot.create(:oauth_application)

    post "/oauth/revoke", params: {
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    expect(response.body).to eq("{}")
  end

  it "client_idパラメータが含まれていないとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      resource_owner_id: user.id,
      application: oauth_application,
      scopes: "read")

    post "/oauth/revoke", params: {
      token: access_token.token,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:forbidden)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("unauthorized_client")
  end

  it "client_secretパラメータが含まれていないとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      resource_owner_id: user.id,
      application: oauth_application,
      scopes: "read")

    post "/oauth/revoke", params: {
      token: access_token.token,
      client_id: oauth_application.uid
    }

    expect(response).to have_http_status(:forbidden)
    json = JSON.parse(response.body)
    expect(json["error"]).to eq("unauthorized_client")
  end

  it "既に無効化されたトークンを送信したとき、成功すること" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      resource_owner_id: user.id,
      application: oauth_application,
      scopes: "read",
      revoked_at: Time.current)

    post "/oauth/revoke", params: {
      token: access_token.token,
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    expect(response.body).to eq("{}")
  end

  it "token_type_hintパラメータでaccess_tokenを指定したとき、正しく動作すること" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      resource_owner_id: user.id,
      application: oauth_application,
      scopes: "read")

    post "/oauth/revoke", params: {
      token: access_token.token,
      token_type_hint: "access_token",
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    expect(response.body).to eq("{}")

    # トークンが無効化されていることを確認
    access_token.reload
    expect(access_token.revoked_at).to be_present
  end

  it "token_type_hintパラメータでrefresh_tokenを指定したとき、正しく動作すること" do
    user = FactoryBot.create(:user)
    oauth_application = FactoryBot.create(:oauth_application)
    access_token = FactoryBot.create(:oauth_access_token,
      resource_owner_id: user.id,
      application: oauth_application,
      scopes: "read",
      refresh_token: "refresh_token_123")

    post "/oauth/revoke", params: {
      token: access_token.refresh_token,
      token_type_hint: "refresh_token",
      client_id: oauth_application.uid,
      client_secret: oauth_application.secret
    }

    expect(response).to have_http_status(:ok)
    expect(response.body).to eq("{}")

    # トークンが無効化されていることを確認
    access_token.reload
    expect(access_token.revoked_at).to be_present
  end
end
