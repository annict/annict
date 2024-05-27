# typed: false
# frozen_string_literal: true

describe "GET /db/episodes/:id/edit", type: :request do
  context "user does not sign in" do
    let!(:episode) { create(:episode) }

    it "user can not access this page" do
      get "/db/episodes/#{episode.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:episode) { create(:episode) }

    before do
      login_as(user, scope: :user)
    end

    it "can not access" do
      get "/db/episodes/#{episode.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:episode) { create(:episode) }

    before do
      login_as(user, scope: :user)
    end

    it "responses episode edit form" do
      get "/db/episodes/#{episode.id}/edit"

      expect(response.status).to eq(200)
      expect(response.body).to include(episode.title)
    end
  end
end
