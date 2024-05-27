# typed: false
# frozen_string_literal: true

describe "POST /db/channels", type: :request do
  context "user does not sign in" do
    let!(:channel_params) do
      {
        name: "ちゃんねる"
      }
    end

    it "user can not access this page" do
      post "/db/channels", params: {channel: channel_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Channel.all.size).to eq(220)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:channel_params) do
      {
        name: "ちゃんねる"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/channels", params: {channel: channel_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Channel.all.size).to eq(220)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:channel_params) do
      {
        name: "ちゃんねる"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/channels", params: {channel: channel_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Channel.all.size).to eq(220)
    end
  end

  context "user who is admin signs in" do
    let!(:channel_group) { ChannelGroup.first }
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:channel_params) do
      {
        channel_group_id: channel_group.id,
        name: "ちゃんねる"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create channel" do
      expect(Channel.all.size).to eq(220)

      post "/db/channels", params: {channel: channel_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Channel.all.size).to eq(221)
      channel = Channel.last

      expect(channel.channel_group_id).to eq(channel_group.id)
      expect(channel.name).to eq("ちゃんねる")
    end
  end
end
