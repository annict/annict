# typed: false
# frozen_string_literal: true

describe "GET /db/organizations", type: :request do
  context "user does not sign in" do
    let!(:organization) { create(:organization) }

    it "responses organization list" do
      get "/db/organizations"

      expect(response.status).to eq(200)
      expect(response.body).to include(organization.name)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:organization) { create(:organization) }

    before do
      login_as(user, scope: :user)
    end

    it "responses organization list" do
      get "/db/organizations"

      expect(response.status).to eq(200)
      expect(response.body).to include(organization.name)
    end
  end
end
