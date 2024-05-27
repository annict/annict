# typed: false
# frozen_string_literal: true

describe "GET /db/works/:work_id/episodes", type: :request do
  context "user does not sign in" do
    let!(:episode) { create(:episode) }

    it "responses episode list" do
      get "/db/works/#{episode.work_id}/episodes"

      expect(response.status).to eq(200)
      expect(response.body).to include(episode.title)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:episode) { create(:episode) }

    before do
      login_as(user, scope: :user)
    end

    it "responses work list" do
      get "/db/works/#{episode.work_id}/episodes"

      expect(response.status).to eq(200)
      expect(response.body).to include(episode.title)
    end
  end
end
