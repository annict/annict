# frozen_string_literal: true

describe "DELETE /db/episodes/:id", type: :request do
  context "user does not sign in" do
    let!(:episode) { create(:episode, :not_deleted) }

    it "user can not access this page" do
      delete "/db/episodes/#{episode.id}"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(episode.deleted?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:episode) { create(:episode, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/episodes/#{episode.id}"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(episode.deleted?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:episode) { create(:episode, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/episodes/#{episode.id}"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(episode.deleted?).to eq(false)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:episode) { create(:episode, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can delete episode softly" do
      expect(episode.deleted?).to eq(false)

      delete "/db/episodes/#{episode.id}"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("削除しました")

      expect(episode.deleted?).to eq(true)
    end
  end
end
