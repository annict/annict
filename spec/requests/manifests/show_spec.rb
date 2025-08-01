# typed: false
# frozen_string_literal: true

RSpec.describe "GET /manifest", type: :request do
  it "JSONをリクエストしたときPWAのマニフェストデータを返すこと" do
    get "/manifest.json"

    expect(response.status).to eq(200)

    actual = JSON.parse(response.body)
    expect(actual["name"]).to eq("Annict")
  end

  it "JSONをリクエストしなかったときは404を返すこと" do
    get "/manifest"

    expect(response.status).to eq(404)
  end
end
