# typed: false
# frozen_string_literal: true

describe "GET /db/people", type: :request do
  context "user does not sign in" do
    let!(:person) { create(:person) }

    it "responses person list" do
      get "/db/people"

      expect(response.status).to eq(200)
      expect(response.body).to include(person.name)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:person) { create(:person) }

    before do
      login_as(user, scope: :user)
    end

    it "responses person list" do
      get "/db/people"

      expect(response.status).to eq(200)
      expect(response.body).to include(person.name)
    end
  end
end
