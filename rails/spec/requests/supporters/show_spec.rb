# typed: false
# frozen_string_literal: true

RSpec.describe "GET /supporters", type: :request do
  it "サポーターページが正常に表示されること" do
    get "/supporters"

    expect(response.status).to eq(200)
    expect(response.body).to include("Annict Supporters")
  end

  it "ページタイトルが適切に設定されていること" do
    get "/supporters"

    expect(response.status).to eq(200)
    expect(response.body).to include("<title>Annict Supporters")
  end
end
