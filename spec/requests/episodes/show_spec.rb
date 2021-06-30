# frozen_string_literal: true

describe "GET /works/:anime_id/episodes/:episode_id", type: :request do
  context "ログインしていないとき" do
    let!(:anime) { create(:anime) }
    let!(:episode) { create(:episode, anime: anime) }

    it "エピソードページが表示されること" do
      get "/works/#{anime.id}/episodes/#{episode.id}"

      expect(response.status).to eq(200)
      expect(response.body).to include(episode.title)
    end
  end

  context "ログインしているとき" do
    let!(:user) { create(:registered_user) }
    let!(:anime) { create(:anime) }
    let!(:episode) { create(:episode, anime: anime) }

    before do
      login_as(user, scope: :user)
    end

    it "エピソードページが表示されること" do
      get "/works/#{anime.id}/episodes/#{episode.id}"

      expect(response.status).to eq(200)
      expect(response.body).to include(anime.title)
    end
  end
end
