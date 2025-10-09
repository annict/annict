# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/:work_id/related_works", type: :request do
  it "ユーザーがログインしていないとき、関連作品が表示されること" do
    work = create(:work)
    series = create(:series)
    create(:series_work, series: series, work: work)
    related_work = create(:work, season_year: 2023, season_name: "spring")
    create(:series_work, series: series, work: related_work)

    get "/works/#{work.id}/related_works"

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
    expect(response.body).to include(related_work.title)
  end

  it "ユーザーがログインしているとき、関連作品が表示されること" do
    user = create(:registered_user)
    work = create(:work)
    series = create(:series)
    create(:series_work, series: series, work: work)
    related_work = create(:work, season_year: 2023, season_name: "spring")
    create(:series_work, series: series, work: related_work)
    login_as(user, scope: :user)

    get "/works/#{work.id}/related_works"

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
    expect(response.body).to include(related_work.title)
  end

  it "存在しない作品にアクセスしたとき、RecordNotFoundエラーが発生すること" do
    expect {
      get "/works/999999/related_works"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除された作品にアクセスしたとき、RecordNotFoundエラーが発生すること" do
    work = create(:work, deleted_at: Time.current)

    expect {
      get "/works/#{work.id}/related_works"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "複数のシリーズに属している作品の場合、すべてのシリーズと関連作品が表示されること" do
    work = create(:work)
    series1 = create(:series, name: "シリーズ1")
    series2 = create(:series, name: "シリーズ2")
    create(:series_work, series: series1, work: work)
    create(:series_work, series: series2, work: work)

    related_work1 = create(:work, title: "関連作品1", season_year: 2023, season_name: "spring")
    related_work2 = create(:work, title: "関連作品2", season_year: 2023, season_name: "summer")
    create(:series_work, series: series1, work: related_work1)
    create(:series_work, series: series2, work: related_work2)

    get "/works/#{work.id}/related_works"

    expect(response.status).to eq(200)
    expect(response.body).to include(series1.name)
    expect(response.body).to include(series2.name)
    expect(response.body).to include(related_work1.title)
    expect(response.body).to include(related_work2.title)
  end

  it "シリーズに属していない作品の場合、関連作品がないことが表示されること" do
    work = create(:work)

    get "/works/#{work.id}/related_works"

    expect(response.status).to eq(200)
  end

  it "削除されたシリーズは表示されないこと" do
    work = create(:work)
    series = create(:series, deleted_at: Time.current)
    create(:series_work, series: series, work: work)

    get "/works/#{work.id}/related_works"

    expect(response.status).to eq(200)
    expect(response.body).not_to include(series.name)
  end

  it "削除された関連作品も表示されること" do
    work = create(:work)
    series = create(:series)
    create(:series_work, series: series, work: work)
    related_work = create(:work, deleted_at: Time.current)
    create(:series_work, series: series, work: related_work)

    get "/works/#{work.id}/related_works"

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
    # 現在の実装では削除された作品も表示される
    expect(response.body).to include(related_work.title)
  end

  it "関連作品がシーズン順（年→季節）でソートされて表示されること" do
    work = create(:work)
    series = create(:series)
    create(:series_work, series: series, work: work)

    work_2024_summer = create(:work, title: "2024夏", season_year: 2024, season_name: "summer")
    work_2023_spring = create(:work, title: "2023春", season_year: 2023, season_name: "spring")
    work_2024_winter = create(:work, title: "2024冬", season_year: 2024, season_name: "winter")

    create(:series_work, series: series, work: work_2024_summer)
    create(:series_work, series: series, work: work_2023_spring)
    create(:series_work, series: series, work: work_2024_winter)

    get "/works/#{work.id}/related_works"

    expect(response.status).to eq(200)
    # 年→季節順でソートされているか確認
    body_index_2023_spring = response.body.index("2023春")
    body_index_2024_winter = response.body.index("2024冬")
    body_index_2024_summer = response.body.index("2024夏")

    expect(body_index_2023_spring).to be < body_index_2024_winter
    expect(body_index_2024_winter).to be < body_index_2024_summer
  end

  it "非公開のシリーズ作品が表示されないこと" do
    series = create(:series)

    published_on_series_work = create(:work)
    unpublished_on_series_work = create(:work)
    create(:series_work, series: series, work: published_on_series_work)
    create(:series_work, series: series, work: unpublished_on_series_work, unpublished_at: Time.current)

    get "/works/#{published_on_series_work.id}/related_works"

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
    expect(response.body).to include(published_on_series_work.title)
    expect(response.body).not_to include(unpublished_on_series_work.title)
  end

  it "非公開でシリーズ作品が登録されている時にシリーズ自体表示されないこと" do
    series = create(:series)
    work = create(:work)
    create(:series_work, series: series, work: work, unpublished_at: Time.current)

    get "/works/#{work.id}/related_works"

    expect(response.status).to eq(200)
    expect(response.body).not_to include("#{series.name} #{I18n.t("noun.series")}")
  end
end
