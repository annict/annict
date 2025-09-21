# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/series/:series_id/series_works", type: :request do
  it "ログインしていないとき、シリーズワーク一覧が表示されること" do
    series_work = FactoryBot.create(:series_work)
    series = series_work.series
    work = series_work.work

    get "/db/series/#{series.id}/series_works"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "ログインしているとき、シリーズワーク一覧が表示されること" do
    user = FactoryBot.create(:registered_user)
    series_work = FactoryBot.create(:series_work)
    series = series_work.series
    work = series_work.work

    login_as(user, scope: :user)

    get "/db/series/#{series.id}/series_works"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "存在しないシリーズIDの場合、404エラーが返されること" do
    non_existent_id = "non-existent-id"

    get "/db/series/#{non_existent_id

    expect(response.status).to eq(404)
  end

  it "削除されたシリーズのワーク一覧にはアクセスできないこと" do
    series_work = FactoryBot.create(:series_work)
    series = series_work.series
    series.destroy!

    get "/db/series/#{series.id

    expect(response.status).to eq(404)
  end

  it "複数のワークが存在する場合、すべて表示されること" do
    series = FactoryBot.create(:series)
    work1 = FactoryBot.create(:work, title: "テストワーク1")
    work2 = FactoryBot.create(:work, title: "テストワーク2")
    FactoryBot.create(:series_work, series:, work: work1)
    FactoryBot.create(:series_work, series:, work: work2)

    get "/db/series/#{series.id}/series_works"

    expect(response.status).to eq(200)
    expect(response.body).to include("テストワーク1")
    expect(response.body).to include("テストワーク2")
  end
end
