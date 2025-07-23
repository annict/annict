# typed: false
# frozen_string_literal: true

RSpec.describe "GET /privacy", type: :request do
  it "プライバシーポリシーページが表示されること" do
    get "/privacy"

    expect(response.status).to eq(200)
    expect(response.body).to include("収集する利用者情報および収集方法")
  end
end
