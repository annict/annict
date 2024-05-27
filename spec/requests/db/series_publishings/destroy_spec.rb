# typed: false
# frozen_string_literal: true

describe "DELETE /db/series/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:series) { create(:series, :published) }

    it "user can not access this page" do
      delete "/db/series/#{series.id}/publishing"
      series.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(series.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:series) { create(:series, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/series/#{series.id}/publishing"
      series.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(series.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:series) { create(:series, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish series" do
      expect(series.published?).to eq(true)

      delete "/db/series/#{series.id}/publishing"
      series.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(series.published?).to eq(false)
    end
  end
end
