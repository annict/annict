# frozen_string_literal: true

describe "PATCH /db/channels/:id", type: :request do
  context "user does not sign in" do
    let!(:channel) { Channel.first }
    let!(:old_channel) { channel.attributes }
    let!(:channel_params) do
      {
        name: "ちゃんねる"
      }
    end

    it "user can not access this page" do
      patch "/db/channels/#{channel.id}", params: {channel: channel_params}
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(channel.name).to eq(old_channel["name"])
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:channel) { Channel.first }
    let!(:old_channel) { channel.attributes }
    let!(:channel_params) do
      {
        name: "ちゃんねる"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/channels/#{channel.id}", params: {channel: channel_params}
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(channel.name).to eq(old_channel["name"])
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:channel) { Channel.first }
    let!(:old_channel) { channel.attributes }
    let!(:channel_params) do
      {
        name: "ちゃんねる"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/channels/#{channel.id}", params: {channel: channel_params}
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(channel.name).to eq(old_channel["name"])
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:channel) { Channel.first }
    let!(:old_channel) { channel.attributes }
    let!(:channel_params) do
      {
        name: "ちゃんねる"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can update channel" do
      expect(channel.name).to eq(old_channel["name"])

      patch "/db/channels/#{channel.id}", params: {channel: channel_params}
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(channel.name).to eq("ちゃんねる")
    end
  end
end
