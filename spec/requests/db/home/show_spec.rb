# typed: false
# frozen_string_literal: true

describe "GET /db", type: :request do
  context "user does not sign in" do
    it "responses the home page" do
      get "/db"

      expect(response.status).to eq(200)
      expect(response.body).to include("Annict DBにようこそ！")
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    it "responses the home page" do
      get "/db"

      expect(response.status).to eq(200)
      expect(response.body).to include("Annict DBにようこそ！")
    end
  end
end
