# typed: false
# frozen_string_literal: true

describe "GET /db/channels", type: :request do
  context "user does not sign in" do
    let!(:channel) { Channel.first }

    it "responses channel list" do
      get "/db/channels"

      expect(response.status).to eq(200)
      expect(response.body).to include(channel.name)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:channel) { Channel.first }

    before do
      login_as(user, scope: :user)
    end

    it "responses channel list" do
      get "/db/channels"

      expect(response.status).to eq(200)
      expect(response.body).to include(channel.name)
    end
  end
end
