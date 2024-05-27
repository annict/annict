# typed: false
# frozen_string_literal: true

describe "PATCH /db/episodes/:id", type: :request do
  context "user does not sign in" do
    let!(:episode) { create(:episode) }
    let!(:old_episode) { episode.attributes }
    let!(:episode_params) do
      {
        title: "タイトルUpdated"
      }
    end

    it "user can not access this page" do
      patch "/db/episodes/#{episode.id}", params: {episode: episode_params}
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(episode.title).to eq(old_episode["title"])
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:episode) { create(:episode) }
    let!(:old_episode) { episode.attributes }
    let!(:episode_params) do
      {
        title: "タイトルUpdated"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/episodes/#{episode.id}", params: {episode: episode_params}
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(episode.title).to eq(old_episode["title"])
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:episode) { create(:episode) }
    let!(:old_episode) { episode.attributes }
    let!(:episode_params) do
      {
        title: "タイトルUpdated"
      }
    end
    let!(:attr_names) { episode_params.keys }

    before do
      login_as(user, scope: :user)
    end

    it "user can update episode" do
      expect(episode.title).to eq(old_episode["title"])

      patch "/db/episodes/#{episode.id}", params: {episode: episode_params}
      episode.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(episode.title).to eq("タイトルUpdated")
    end
  end
end
