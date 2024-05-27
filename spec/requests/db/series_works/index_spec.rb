# typed: false
# frozen_string_literal: true

describe "GET /db/series/:series_id/series_works", type: :request do
  context "user does not sign in" do
    let!(:series_work) { create(:series_work) }
    let!(:series) { series_work.series }
    let!(:work) { series_work.work }

    it "responses series work list" do
      get "/db/series/#{series.id}/series_works"

      expect(response.status).to eq(200)
      expect(response.body).to include(work.title)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:series_work) { create(:series_work) }
    let!(:series) { series_work.series }
    let!(:work) { series_work.work }

    before do
      login_as(user, scope: :user)
    end

    it "responses series list" do
      get "/db/series/#{series.id}/series_works"

      expect(response.status).to eq(200)
      expect(response.body).to include(work.title)
    end
  end
end
