# frozen_string_literal: true

describe "GET /works/:season_slug", type: :request do
  let!(:anime) { create(:work, :with_current_season) }
  let(:current_season_slug) { ENV["ANNICT_CURRENT_SEASON"] }

  before do
    host! ENV.fetch("ANNICT_JP_HOST")
  end

  it "アクセスできること" do
    get "/works/#{current_season_slug}"

    expect(response.status).to eq(200)
    expect(response.body).to include(anime.title)
  end
end
