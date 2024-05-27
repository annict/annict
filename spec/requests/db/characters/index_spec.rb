# typed: false
# frozen_string_literal: true

describe "GET /db/characters", type: :request do
  context "user does not sign in" do
    let!(:character) { create(:character) }

    it "responses character list" do
      get "/db/characters"

      expect(response.status).to eq(200)
      expect(response.body).to include(character.name)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:character) { create(:character) }

    before do
      login_as(user, scope: :user)
    end

    it "responses character list" do
      get "/db/characters"

      expect(response.status).to eq(200)
      expect(response.body).to include(character.name)
    end
  end
end
