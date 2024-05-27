# typed: false
# frozen_string_literal: true

describe "GET /supporters", type: :request do
  it "アクセスできること" do
    get "/supporters"

    expect(response.status).to eq(200)
    expect(response.body).to include("Annictサポーター")
  end
end
