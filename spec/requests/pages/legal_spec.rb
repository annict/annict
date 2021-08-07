# frozen_string_literal: true

describe "GET /legal", type: :request do
  it "特商法ページが表示されること" do
    get "/legal"

    expect(response.status).to eq(200)
    expect(response.body).to include("サービス名")
  end
end
