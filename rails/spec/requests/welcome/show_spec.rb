# typed: false
# frozen_string_literal: true

RSpec.describe "GET /", type: :request do
  it "ログインしていないとき、アニメが登録されていないとき、Welcomeページが表示されること" do
    get "/"

    expect(response.status).to eq(200)
    expect(response.body).to include("A platform for anime addicts.")
    expect(response.body).to include("作品はありません")
  end

  it "ログインしていないとき、アニメが登録されているとき、Welcomeページが表示されること" do
    work = FactoryBot.create(:work, :with_current_season)

    get "/"

    expect(response.status).to eq(200)
    expect(response.body).to include("A platform for anime addicts.")
    expect(response.body).to include(work.title)
  end
end
