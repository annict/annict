# typed: false
# frozen_string_literal: true

describe "GET /legacy/sign_in", type: :request do
  it "パスワードでログインページが表示されること" do
    get "/legacy/sign_in"

    expect(response.status).to eq(200)
    expect(response.body).to include("おかえりなさい！")
  end
end
