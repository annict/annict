# frozen_string_literal: true

describe "GET /works/:slug", type: :request do
  context "アニメが登録されていないとき" do
    it "アクセスできること" do
      get "/works/#{ENV["ANNICT_CURRENT_SEASON"]}"

      expect(response.status).to eq(200)
      expect(response.body).to include("2017年冬のアニメ")
    end
  end

  context "アニメが登録されているとき" do
    let!(:anime) { create(:anime, :with_current_season) }

    it "アクセスできること" do
      get "/works/#{ENV["ANNICT_CURRENT_SEASON"]}"

      expect(response.status).to eq(200)
      expect(response.body).to include(anime.title)
    end
  end
end
