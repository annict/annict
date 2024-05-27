# typed: false
# frozen_string_literal: true

describe "GET /db/series", type: :request do
  context "user does not sign in" do
    let!(:series) { create(:series) }

    it "responses series list" do
      get "/db/series"

      expect(response.status).to eq(200)
      expect(response.body).to include(series.name)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:series) { create(:series) }

    before do
      login_as(user, scope: :user)
    end

    it "responses series list" do
      get "/db/series"

      expect(response.status).to eq(200)
      expect(response.body).to include(series.name)
    end
  end
end
