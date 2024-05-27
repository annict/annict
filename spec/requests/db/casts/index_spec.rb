# typed: false
# frozen_string_literal: true

describe "GET /db/works/:work_id/casts", type: :request do
  context "user does not sign in" do
    let!(:cast) { create(:cast) }

    it "responses cast list" do
      get "/db/works/#{cast.work_id}/casts"

      expect(response.status).to eq(200)
      expect(response.body).to include(cast.character.name)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:cast) { create(:cast) }

    before do
      login_as(user, scope: :user)
    end

    it "responses work list" do
      get "/db/works/#{cast.work_id}/casts"

      expect(response.status).to eq(200)
      expect(response.body).to include(cast.character.name)
    end
  end
end
