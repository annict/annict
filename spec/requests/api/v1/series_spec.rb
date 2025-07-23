# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/series", type: :request do
  it "パラメーターを指定しない場合、200ステータスを返すこと" do
    access_token = create(:oauth_access_token)
    create(:series)

    get api("/v1/series", access_token: access_token.token)

    expect(response.status).to eq(200)
  end

  it "パラメーターを指定しない場合、シリーズ情報を取得できること" do
    access_token = create(:oauth_access_token)
    series = create(:series)

    get api("/v1/series", access_token: access_token.token)

    expected_hash = {
      "id" => series.id,
      "name" => series.name,
      "name_ro" => series.name_ro,
      "name_en" => series.name_en
    }
    expect(json["series"][0]).to include(expected_hash)
    expect(json["total_count"]).to eq(1)
    expect(json["next_page"]).to eq(nil)
    expect(json["prev_page"]).to eq(nil)
  end

  it "アクセストークンが無効な場合、401ステータスを返すこと" do
    get api("/v1/series", access_token: "invalid_token")

    expect(response.status).to eq(401)
  end

  it "シリーズが存在しない場合、空の配列を返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/series", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["series"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end
end
