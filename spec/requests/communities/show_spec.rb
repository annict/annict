# typed: false
# frozen_string_literal: true

RSpec.describe "GET /community", type: :request do
  it "環境変数ANNICT_COMMUNITY_URLが設定されているとき、コミュニティURLにリダイレクトされること" do
    community_url = "https://discord.gg/annict"
    allow(ENV).to receive(:fetch).and_call_original
    allow(ENV).to receive(:fetch).with("ANNICT_COMMUNITY_URL").and_return(community_url)

    get "/community"

    expect(response.status).to eq(302)
    expect(response.location).to eq(community_url)
  end

  it "環境変数ANNICT_COMMUNITY_URLが設定されていないとき、KeyErrorが発生すること" do
    allow(ENV).to receive(:fetch).and_call_original
    allow(ENV).to receive(:fetch).with("ANNICT_COMMUNITY_URL").and_raise(KeyError)

    expect { get "/community" }.to raise_error(KeyError)
  end
end
