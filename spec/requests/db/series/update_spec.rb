# typed: false
# frozen_string_literal: true

describe "PATCH /db/series/:id", type: :request do
  context "user does not sign in" do
    let!(:series) { create(:series) }
    let!(:old_series) { series.attributes }
    let!(:series_params) do
      {
        name: "シリーズ2",
        name_alter: "シリーズ2 (別名)",
        name_en: "The Series2",
        name_alter_en: "The Series2 (alt)"
      }
    end

    it "user can not access this page" do
      patch "/db/series/#{series.id}", params: {series: series_params}
      series.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(series.send(:name)).to eq(old_series["name"])
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:series) { create(:series) }
    let!(:old_series) { series.attributes }
    let!(:series_params) do
      {
        name: "シリーズ2",
        name_alter: "シリーズ2 (別名)",
        name_en: "The Series2",
        name_alter_en: "The Series2 (alt)"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/series/#{series.id}", params: {series: series_params}
      series.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(series.send(:name)).to eq(old_series["name"])
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:series) { create(:series) }
    let!(:old_series) { series.attributes }
    let!(:attr_names) do
      %i[
        name
        name_alter
        name_en
        name_alter_en
      ]
    end
    let!(:series_params) do
      {
        name: "シリーズ2",
        name_alter: "シリーズ2 (別名)",
        name_en: "The Series2",
        name_alter_en: "The Series2 (alt)"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can update series" do
      attr_names.each do |attr_name|
        expect(series.send(attr_name)).to eq(old_series[attr_name.to_s])
      end

      patch "/db/series/#{series.id}", params: {series: series_params}
      series.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      attr_names.each do |attr_name|
        expect(series.send(attr_name)).to eq(series_params[attr_name])
      end
    end
  end
end
