# typed: false
# frozen_string_literal: true

describe "PATCH /db/channel_groups/:id", type: :request do
  context "user does not sign in" do
    let!(:channel_group) { ChannelGroup.first }
    let!(:old_channel_group) { channel_group.attributes }
    let!(:channel_group_params) do
      {
        name: "ちゃんねるぐるーぷ"
      }
    end

    it "user can not access this page" do
      patch "/db/channel_groups/#{channel_group.id}", params: {channel_group: channel_group_params}
      channel_group.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(channel_group.name).to eq(old_channel_group["name"])
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:channel_group) { ChannelGroup.first }
    let!(:old_channel_group) { channel_group.attributes }
    let!(:channel_group_params) do
      {
        name: "ちゃんねるぐるーぷ"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/channel_groups/#{channel_group.id}", params: {channel_group: channel_group_params}
      channel_group.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(channel_group.name).to eq(old_channel_group["name"])
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:channel_group) { ChannelGroup.first }
    let!(:old_channel_group) { channel_group.attributes }
    let!(:channel_group_params) do
      {
        name: "ちゃんねるぐるーぷ"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/channel_groups/#{channel_group.id}", params: {channel_group: channel_group_params}
      channel_group.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(channel_group.name).to eq(old_channel_group["name"])
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:channel_group) { ChannelGroup.first }
    let!(:old_channel_group) { channel_group.attributes }
    let!(:channel_group_params) do
      {
        name: "ちゃんねるぐるーぷ"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can update channel_group" do
      expect(channel_group.name).to eq(old_channel_group["name"])

      patch "/db/channel_groups/#{channel_group.id}", params: {channel_group: channel_group_params}
      channel_group.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(channel_group.name).to eq("ちゃんねるぐるーぷ")
    end
  end
end
