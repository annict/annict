# frozen_string_literal: true

describe "GET /works/newest", type: :request do
  let!(:anime) { create(:work) }

  before do
    host! ENV.fetch("ANNICT_JP_HOST")
  end

  it "アクセスできること" do
    get "/works/newest"

    expect(response.status).to eq(200)
    expect(response.body).to include(anime.title)
  end
end
