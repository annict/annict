# frozen_string_literal: true

describe "GET /works/popular", type: :request do
  let!(:work) { create(:work) }

  it "アクセスできること" do
    get "/works/popular"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end
end
