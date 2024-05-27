# typed: false
# frozen_string_literal: true

describe "GET /db/organizations/:id/edit", type: :request do
  context "user does not sign in" do
    let!(:organization) { create(:organization) }

    it "user can not access this page" do
      get "/db/organizations/#{organization.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:organization) { create(:organization) }

    before do
      login_as(user, scope: :user)
    end

    it "can not access" do
      get "/db/organizations/#{organization.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:organization) { create(:organization) }

    before do
      login_as(user, scope: :user)
    end

    it "responses organization edit form" do
      get "/db/organizations/#{organization.id}/edit"

      expect(response.status).to eq(200)
      expect(response.body).to include(organization.name)
    end
  end
end
