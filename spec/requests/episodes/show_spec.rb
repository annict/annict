# frozen_string_literal: true

describe "GET /works/:work_id/episodes/:episode_id", type: :request do
  context "ログインしていないとき" do
    let!(:work) { create(:work) }
    let!(:episode) { create(:episode, work: work) }

    it "エピソードページが表示されること" do
      get "/works/#{work.id}/episodes/#{episode.id}"

      expect(response.status).to eq(200)
      expect(response.body).to include(episode.title)
    end
  end

  context "ログインしているとき" do
    let!(:user) { create(:registered_user) }
    let!(:work) { create(:work) }
    let!(:episode) { create(:episode, work: work) }

    before do
      login_as(user, scope: :user)
    end

    it "エピソードページが表示されること" do
      get "/works/#{work.id}/episodes/#{episode.id}"

      expect(response.status).to eq(200)
      expect(response.body).to include(work.title)
    end
  end
end
