# typed: false
# frozen_string_literal: true

describe "GET /sign_up", type: :request do
  it "サインアップページが表示されること" do
    get "/sign_up"

    expect(response.status).to eq(200)
    expect(response.body).to include("アカウント作成")
  end
end
