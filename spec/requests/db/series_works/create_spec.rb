# frozen_string_literal: true

describe "POST /db/series/:series_id/series_works", type: :request do
  context "user does not sign in" do
    let!(:series) { create(:series) }
    let!(:work) { create(:anime) }
    let!(:form_params) do
      {
        rows: "#{work.id}, Season 1"
      }
    end

    it "user can not access this page" do
      post "/db/series/#{series.id}/series_works", params: {db_series_work_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(SeriesAnime.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:series) { create(:series) }
    let!(:work) { create(:anime) }
    let!(:form_params) do
      {
        rows: "#{work.id}, Season 1"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/series/#{series.id}/series_works", params: {db_series_work_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(SeriesAnime.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:series) { create(:series) }
    let!(:work) { create(:anime) }
    let!(:form_params) do
      {
        rows: "#{work.id}, Season 1"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create series" do
      expect(SeriesAnime.all.size).to eq(0)

      post "/db/series/#{series.id}/series_works", params: {db_series_work_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(SeriesAnime.all.size).to eq(1)
      series_work = SeriesAnime.first
      expect(series_work.anime).to eq(work)
      expect(series_work.summary).to eq("Season 1")
    end
  end
end
