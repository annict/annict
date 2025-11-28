# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/*path", type: :request do
  it "存在しないパスにアクセスした場合、404ステータスを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/nonexistent", access_token: access_token.token)

    expect(response.status).to eq(404)
  end

  it "存在しないパスにアクセスした場合、エラーメッセージを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/nonexistent", access_token: access_token.token)

    expect(json["errors"]).to be_present
    expect(json["errors"].first["type"]).to eq("unknown_route")
    expect(json["errors"].first["message"]).to eq("リクエストに失敗しました")
    expect(json["errors"].first["developer_message"]).to eq("404 Not Found")
  end

  it "存在しないパスにPOSTでアクセスした場合、404ステータスを返すこと" do
    access_token = create(:oauth_access_token)

    post api("/v1/nonexistent", access_token: access_token.token)

    expect(response.status).to eq(404)
  end

  it "存在しないパスにPUTでアクセスした場合、404ステータスを返すこと" do
    access_token = create(:oauth_access_token)

    put api("/v1/nonexistent", access_token: access_token.token)

    expect(response.status).to eq(404)
  end

  it "存在しないパスにDELETEでアクセスした場合、404ステータスを返すこと" do
    access_token = create(:oauth_access_token)

    delete api("/v1/nonexistent", access_token: access_token.token)

    expect(response.status).to eq(404)
  end

  it "深いパスにアクセスした場合、404ステータスを返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/nonexistent/deep/path", access_token: access_token.token)

    expect(response.status).to eq(404)
  end

  it "アクセストークンが無効な場合、404ステータスを返すこと" do
    get api("/v1/nonexistent", access_token: "invalid_token")

    expect(response.status).to eq(404)
  end

  it "アクセストークンが無い場合、404ステータスを返すこと" do
    get "/v1/nonexistent"

    expect(response.status).to eq(404)
  end
end
