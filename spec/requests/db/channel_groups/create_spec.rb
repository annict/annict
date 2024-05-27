# typed: false
# frozen_string_literal: true

describe "POST /db/channel_groups", type: :request do
  context "user does not sign in" do
    let!(:channel_group_params) do
      {
        name: "ちゃんねるぐるーぷ"
      }
    end

    it "user can not access this page" do
      post "/db/channel_groups", params: {channel_group: channel_group_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(ChannelGroup.all.size).to eq(18)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:channel_group_params) do
      {
        name: "ちゃんねるぐるーぷ"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/channel_groups", params: {channel_group: channel_group_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(ChannelGroup.all.size).to eq(18)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:channel_group_params) do
      {
        name: "ちゃんねるぐるーぷ"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/channel_groups", params: {channel_group: channel_group_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(ChannelGroup.all.size).to eq(18)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:channel_group_params) do
      {
        name: "ちゃんねるぐるーぷ"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create channel_group" do
      expect(ChannelGroup.all.size).to eq(18)

      post "/db/channel_groups", params: {channel_group: channel_group_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(ChannelGroup.all.size).to eq(19)
      channel_group = ChannelGroup.last

      expect(channel_group.name).to eq("ちゃんねるぐるーぷ")
    end
  end
end
