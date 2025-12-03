# typed: false
# frozen_string_literal: true

RSpec.describe "GET /community", type: :request do
  around do |example|
    original_value = ENV["ANNICT_COMMUNITY_URL"]
    example.run
    ENV["ANNICT_COMMUNITY_URL"] = original_value
  end

  it "環境変数ANNICT_COMMUNITY_URLが設定されているとき、コミュニティURLにリダイレクトされること" do
    community_url = "https://discord.gg/annict"
    ENV["ANNICT_COMMUNITY_URL"] = community_url

    get "/community"

    expect(response.status).to eq(302)
    expect(response.location).to eq(community_url)
  end

  it "環境変数ANNICT_COMMUNITY_URLが設定されていないとき、KeyErrorが発生すること" do
    ENV.delete("ANNICT_COMMUNITY_URL")

    expect { get "/community" }.to raise_error(KeyError)
  end
end
