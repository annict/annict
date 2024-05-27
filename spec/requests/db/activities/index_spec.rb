# typed: false
# frozen_string_literal: true

describe "GET /db/activities", type: :request do
  context "user does not sign in" do
    let!(:db_activity) { create(:works_create_activity) }

    it "responses activity list" do
      work = db_activity.trackable

      get "/db/activities"

      expect(response.status).to eq(200)
      expect(response.body).to include(work.title)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:db_activity) { create(:works_create_activity) }

    before do
      login_as(user, scope: :user)
    end

    it "responses series list" do
      work = db_activity.trackable

      get "/db/activities"

      expect(response.status).to eq(200)
      expect(response.body).to include(work.title)
    end
  end
end
