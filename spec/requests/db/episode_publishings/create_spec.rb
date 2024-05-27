# typed: false
# frozen_string_literal: true

describe "POST /db/episodes/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:episode) { create(:episode, :unpublished) }

    it "user can not access this page" do
      post "/db/episodes/#{episode.id}/publishing"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(episode.published?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:episode) { create(:episode, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/episodes/#{episode.id}/publishing"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(episode.published?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:episode) { create(:episode, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can publish episode" do
      expect(episode.published?).to eq(false)

      post "/db/episodes/#{episode.id}/publishing"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("公開しました")

      expect(episode.published?).to eq(true)
    end
  end
end
