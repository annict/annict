# typed: false
# frozen_string_literal: true

describe "DELETE /db/episodes/:id", type: :request do
  context "user does not sign in" do
    let!(:episode) { create(:episode, :not_deleted) }

    it "user can not access this page" do
      expect(Episode.count).to eq(1)

      delete "/db/episodes/#{episode.id}"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Episode.count).to eq(1)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:episode) { create(:episode, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Episode.count).to eq(1)

      delete "/db/episodes/#{episode.id}"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Episode.count).to eq(1)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:episode) { create(:episode, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Episode.count).to eq(1)

      delete "/db/episodes/#{episode.id}"
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Episode.count).to eq(1)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:episode) { create(:episode, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can delete episode softly" do
      expect(Episode.count).to eq(1)

      delete "/db/episodes/#{episode.id}"

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("削除しました")

      expect(Episode.count).to eq(0)
    end
  end
end
