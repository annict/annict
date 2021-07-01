# frozen_string_literal: true

describe "GET /works/newest", type: :request do
  let!(:anime) { create(:anime) }

  it "アクセスできること" do
    get "/works/newest"

    expect(response.status).to eq(200)
    expect(response.body).to include(anime.title)
  end
end
