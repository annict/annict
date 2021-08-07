# frozen_string_literal: true

describe "GET /db/channel_groups/:id/edit", type: :request do
  context "user does not sign in" do
    let!(:channel_group) { ChannelGroup.first }

    it "user can not access this page" do
      get "/db/channel_groups/#{channel_group.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:channel_group) { ChannelGroup.first }

    before do
      login_as(user, scope: :user)
    end

    it "can not access" do
      get "/db/channel_groups/#{channel_group.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:channel_group) { ChannelGroup.first }

    before do
      login_as(user, scope: :user)
    end

    it "can not access" do
      get "/db/channel_groups/#{channel_group.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:channel_group) { ChannelGroup.first }

    before do
      login_as(user, scope: :user)
    end

    it "responses channel_group edit form" do
      get "/db/channel_groups/#{channel_group.id}/edit"

      expect(response.status).to eq(200)
      expect(response.body).to include(channel_group.name)
    end
  end
end
