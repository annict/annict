# typed: false
# frozen_string_literal: true

describe "PATCH /db/series_works/:id", type: :request do
  context "user does not sign in" do
    let!(:series_work) { create(:series_work) }
    let!(:old_series_work) { series_work.attributes }
    let!(:series_work_params) do
      {
        work_id: series_work.work_id,
        summary: "2期",
        summary_en: "Season 2"
      }
    end

    it "user can not access this page" do
      patch "/db/series_works/#{series_work.id}", params: {series_work: series_work_params}
      series_work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(series_work.summary).to eq(old_series_work["summary"])
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:series_work) { create(:series_work) }
    let!(:old_series_work) { series_work.attributes }
    let!(:series_work_params) do
      {
        work_id: series_work.work_id,
        summary: "2期",
        summary_en: "Season 2"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/series_works/#{series_work.id}", params: {series_work: series_work_params}
      series_work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(series_work.summary).to eq(old_series_work["summary"])
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:series_work) { create(:series_work) }
    let!(:old_series_work) { series_work.attributes }
    let!(:series_work_params) do
      {
        work_id: series_work.work_id,
        summary: "2期",
        summary_en: "Season 2"
      }
    end
    let!(:attr_names) do
      %i[
        work_id
        summary
        summary_en
      ]
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can update series work" do
      attr_names.each do |attr_name|
        expect(series_work.send(attr_name)).to eq(old_series_work[attr_name.to_s])
      end

      patch "/db/series_works/#{series_work.id}", params: {series_work: series_work_params}
      series_work.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      attr_names.each do |attr_name|
        expect(series_work.send(attr_name)).to eq(series_work_params[attr_name])
      end
    end
  end
end
