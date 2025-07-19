# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db", type: :request do
  it "ユーザーがログインしていないとき、ホームページが表示されること" do
    get "/db"

    expect(response.status).to eq(200)
    expect(response.body).to include("Annict DBにようこそ！")
  end

  it "ユーザーがログインしているとき、ホームページが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/db"

    expect(response.status).to eq(200)
    expect(response.body).to include("Annict DBにようこそ！")
  end
end
