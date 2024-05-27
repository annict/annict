# typed: false
# frozen_string_literal: true

describe "GET /db/casts/:id/edit", type: :request do
  context "user does not sign in" do
    let!(:cast) { create(:cast) }

    it "user can not access this page" do
      get "/db/casts/#{cast.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:cast) { create(:cast) }

    before do
      login_as(user, scope: :user)
    end

    it "can not access" do
      get "/db/casts/#{cast.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:cast) { create(:cast) }

    before do
      login_as(user, scope: :user)
    end

    it "responses cast edit form" do
      get "/db/casts/#{cast.id}/edit"

      expect(response.status).to eq(200)
      expect(response.body).to include(cast.character.name)
    end
  end
end
