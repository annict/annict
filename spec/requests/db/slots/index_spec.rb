# typed: false
# frozen_string_literal: true

describe "GET /db/works/:work_id/slots", type: :request do
  context "user does not sign in" do
    let!(:slot) { create(:slot) }

    it "responses slot list" do
      get "/db/works/#{slot.work_id}/slots"

      expect(response.status).to eq(200)
      expect(response.body).to include(slot.channel.name)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:slot) { create(:slot) }

    before do
      login_as(user, scope: :user)
    end

    it "responses work list" do
      get "/db/works/#{slot.work_id}/slots"

      expect(response.status).to eq(200)
      expect(response.body).to include(slot.channel.name)
    end
  end
end
