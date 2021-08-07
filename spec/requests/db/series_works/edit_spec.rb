# frozen_string_literal: true

describe "GET /db/series_works/:id/edit", type: :request do
  context "user does not sign in" do
    let!(:series_work) { create(:series_anime) }

    it "user can not access this page" do
      get "/db/series_works/#{series_work.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:series_work) { create(:series_anime) }

    before do
      login_as(user, scope: :user)
    end

    it "can not access" do
      get "/db/series_works/#{series_work.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:series_work) { create(:series_anime) }
    let!(:work) { series_work.anime }

    before do
      login_as(user, scope: :user)
    end

    it "responses series work form" do
      get "/db/series_works/#{series_work.id}/edit"

      expect(response.status).to eq(200)
      expect(response.body).to include(work.title)
    end
  end
end
