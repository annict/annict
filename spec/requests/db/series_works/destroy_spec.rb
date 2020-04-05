# frozen_string_literal: true

describe "DELETE /db/series_works/:id", type: :request do
  context "user does not sign in" do
    let!(:series_work) { create(:series_work, :not_deleted) }

    it "user can not access this page" do
      delete "/db/series_works/#{series_work.id}"
      series_work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(series_work.deleted?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:series_work) { create(:series_work, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/series_works/#{series_work.id}"
      series_work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(series_work.deleted?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:series_work) { create(:series_work, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/series_works/#{series_work.id}"
      series_work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(series_work.deleted?).to eq(false)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:series_work) { create(:series_work, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can delete series work softly" do
      expect(series_work.deleted?).to eq(false)

      delete "/db/series_works/#{series_work.id}"
      series_work.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("削除しました")

      expect(series_work.deleted?).to eq(true)
    end
  end
end
