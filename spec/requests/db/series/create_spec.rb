# frozen_string_literal: true

describe "POST /db/series", type: :request do
  context "user does not sign in" do
    let!(:series_params) do
      {
        name: "シリーズ",
        name_alter: "シリーズ (別名)",
        name_en: "The Series",
        name_alter_en: "The Series (alt)"
      }
    end

    it "user can not access this page" do
      post "/db/series", params: {series: series_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Series.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:series_params) do
      {
        name: "シリーズ",
        name_alter: "シリーズ (別名)",
        name_en: "The Series",
        name_alter_en: "The Series (alt)"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/series", params: {series: series_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Series.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:series_params) do
      {
        name: "シリーズ",
        name_alter: "シリーズ (別名)",
        name_en: "The Series",
        name_alter_en: "The Series (alt)"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create series" do
      expect(Series.all.size).to eq(0)

      post "/db/series", params: {series: series_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Series.all.size).to eq(1)
      series = Series.first
      expect(series.name).to eq(series_params[:name])
      expect(series.name_alter).to eq(series_params[:name_alter])
      expect(series.name_en).to eq(series_params[:name_en])
      expect(series.name_alter_en).to eq(series_params[:name_alter_en])
    end
  end
end
