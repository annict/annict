# typed: false
# frozen_string_literal: true

describe "GET /db/works/:work_id/trailers", type: :request do
  context "user does not sign in" do
    let!(:trailer) { create(:trailer) }

    it "responses trailer list" do
      get "/db/works/#{trailer.work_id}/trailers"

      expect(response.status).to eq(200)
      expect(response.body).to include(trailer.title)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:trailer) { create(:trailer) }

    before do
      login_as(user, scope: :user)
    end

    it "responses work list" do
      get "/db/works/#{trailer.work_id}/trailers"

      expect(response.status).to eq(200)
      expect(response.body).to include(trailer.title)
    end
  end
end
