# frozen_string_literal: true

describe "GET /db/works", type: :request do
  context "user does not sign in" do
    let!(:work) { create(:anime) }

    it "responses work list" do
      get "/db/works"

      expect(response.status).to eq(200)
      expect(response.body).to include(work.title)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:work) { create(:anime) }

    before do
      login_as(user, scope: :user)
    end

    it "responses work list" do
      get "/db/works"

      expect(response.status).to eq(200)
      expect(response.body).to include(work.title)
    end
  end
end
