# frozen_string_literal: true

describe "POST /db/series_works/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:series_work) { create(:series_anime, :unpublished) }

    it "user can not access this page" do
      post "/db/series_works/#{series_work.id}/publishing"
      series_work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(series_work.published?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:series_work) { create(:series_anime, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/series_works/#{series_work.id}/publishing"
      series_work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(series_work.published?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:series_work) { create(:series_anime, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can publish series work" do
      expect(series_work.published?).to eq(false)

      post "/db/series_works/#{series_work.id}/publishing"
      series_work.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("公開しました")

      expect(series_work.published?).to eq(true)
    end
  end
end
