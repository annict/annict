# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/series/:series_id/series_works", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    series = create(:series)
    work = create(:work)
    form_params = {
      rows: "#{work.id}, Season 1"
    }

    post "/db/series/#{series.id}/series_works", params: {deprecated_db_series_work_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(SeriesWork.all.size).to eq(0)
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    series = create(:series)
    work = create(:work)
    form_params = {
      rows: "#{work.id}, Season 1"
    }

    login_as(user, scope: :user)
    post "/db/series/#{series.id}/series_works", params: {deprecated_db_series_work_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(SeriesWork.all.size).to eq(0)
  end

  it "エディター権限を持つユーザーがログインしているとき、シリーズ作品を作成できること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series)
    work = create(:work)
    form_params = {
      rows: "#{work.id}, Season 1"
    }

    login_as(user, scope: :user)
    expect(SeriesWork.all.size).to eq(0)

    post "/db/series/#{series.id}/series_works", params: {deprecated_db_series_work_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(SeriesWork.all.size).to eq(1)

    series_work = SeriesWork.first
    expect(series_work.work).to eq(work)
    expect(series_work.summary).to eq("Season 1")
  end

  it "エディター権限を持つユーザーがログインしているとき、無効なワークIDで作成に失敗すること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series)
    form_params = {
      rows: "999999, Season 1"
    }

    login_as(user, scope: :user)
    expect(SeriesWork.all.size).to eq(0)

    post "/db/series/#{series.id}/series_works", params: {deprecated_db_series_work_rows_form: form_params}

    expect(response.status).to eq(422)
    expect(SeriesWork.all.size).to eq(0)
  end

  it "エディター権限を持つユーザーがログインしているとき、空のrowsパラメータで作成に失敗すること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series)
    form_params = {
      rows: ""
    }

    login_as(user, scope: :user)
    expect(SeriesWork.all.size).to eq(0)

    post "/db/series/#{series.id}/series_works", params: {deprecated_db_series_work_rows_form: form_params}

    expect(response.status).to eq(422)
    expect(SeriesWork.all.size).to eq(0)
  end

  it "エディター権限を持つユーザーがログインしているとき、複数の作品を一度に作成できること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series)
    work1 = create(:work)
    work2 = create(:work)
    form_params = {
      rows: "#{work1.id}, Season 1\n#{work2.id}, Season 2"
    }

    login_as(user, scope: :user)
    expect(SeriesWork.all.size).to eq(0)

    post "/db/series/#{series.id}/series_works", params: {deprecated_db_series_work_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(SeriesWork.all.size).to eq(2)

    series_works = SeriesWork.all.order(:created_at)
    expect(series_works[0].work).to eq(work1)
    expect(series_works[0].summary).to eq("Season 1")
    expect(series_works[1].work).to eq(work2)
    expect(series_works[1].summary).to eq("Season 2")
  end
end
