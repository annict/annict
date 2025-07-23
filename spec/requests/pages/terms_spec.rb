# typed: false
# frozen_string_literal: true

RSpec.describe "GET /terms", type: :request do
  it "利用規約ページが表示されること" do
    get "/terms"

    expect(response.status).to eq(200)
    expect(response.body).to include("本利用規約")
  end
end
