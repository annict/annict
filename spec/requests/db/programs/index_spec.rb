# typed: false
# frozen_string_literal: true

describe "GET /db/works/:work_id/programs", type: :request do
  context "user does not sign in" do
    let!(:program) { create(:program) }

    it "responses program list" do
      get "/db/works/#{program.work_id}/programs"

      expect(response.status).to eq(200)
      expect(response.body).to include(program.channel.name)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:program) { create(:program) }

    before do
      login_as(user, scope: :user)
    end

    it "responses work list" do
      get "/db/works/#{program.work_id}/programs"

      expect(response.status).to eq(200)
      expect(response.body).to include(program.channel.name)
    end
  end
end
