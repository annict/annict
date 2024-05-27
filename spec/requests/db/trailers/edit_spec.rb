# typed: false
# frozen_string_literal: true

describe "GET /db/trailers/:id/edit", type: :request do
  context "user does not sign in" do
    let!(:trailer) { create(:trailer) }

    it "user can not access this page" do
      get "/db/trailers/#{trailer.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:trailer) { create(:trailer) }

    before do
      login_as(user, scope: :user)
    end

    it "can not access" do
      get "/db/trailers/#{trailer.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:trailer) { create(:trailer) }

    before do
      login_as(user, scope: :user)
    end

    it "responses trailer edit form" do
      get "/db/trailers/#{trailer.id}/edit"

      expect(response.status).to eq(200)
      expect(response.body).to include(trailer.title)
    end
  end
end
