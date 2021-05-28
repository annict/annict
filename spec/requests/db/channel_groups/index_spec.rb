# frozen_string_literal: true

describe "GET /db/channel_groups", type: :request do
  context "user does not sign in" do
    let!(:channel_group) { ChannelGroup.first }

    it "responses channel_group list" do
      get "/db/channel_groups"

      expect(response.status).to eq(200)
      expect(response.body).to include(channel_group.name)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:channel_group) { ChannelGroup.first }

    before do
      login_as(user, scope: :user)
    end

    it "responses channel_group list" do
      get "/db/channel_groups"

      expect(response.status).to eq(200)
      expect(response.body).to include(channel_group.name)
    end
  end
end
