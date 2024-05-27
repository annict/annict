# typed: false
# frozen_string_literal: true

describe "DELETE /db/episodes/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:episode) { create(:episode, :published) }

    it "user can not access this page" do
      delete "/db/episodes/#{episode.id}/publishing"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(episode.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:episode) { create(:episode, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/episodes/#{episode.id}/publishing"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(episode.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:episode) { create(:episode, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish episode" do
      expect(episode.published?).to eq(true)

      delete "/db/episodes/#{episode.id}/publishing"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(episode.published?).to eq(false)
    end
  end
end
