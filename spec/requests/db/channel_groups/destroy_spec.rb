# typed: false
# frozen_string_literal: true

describe "DELETE /db/channel_groups/:id", type: :request do
  context "user does not sign in" do
    let!(:channel_group) { ChannelGroup.first }

    it "user can not access this page" do
      expect(ChannelGroup.count).to eq(18)

      delete "/db/channel_groups/#{channel_group.id}"
      channel_group.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(ChannelGroup.count).to eq(18)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:channel_group) { ChannelGroup.first }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(ChannelGroup.count).to eq(18)

      delete "/db/channel_groups/#{channel_group.id}"
      channel_group.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(ChannelGroup.count).to eq(18)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:channel_group) { ChannelGroup.first }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(ChannelGroup.count).to eq(18)

      delete "/db/channel_groups/#{channel_group.id}"
      channel_group.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(ChannelGroup.count).to eq(18)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:channel_group) { ChannelGroup.first }

    before do
      login_as(user, scope: :user)
    end

    it "user can delete channel_group softly" do
      expect(ChannelGroup.count).to eq(18)

      expect(channel_group.deleted?).to eq(false)

      delete "/db/channel_groups/#{channel_group.id}"

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("削除しました")

      expect(ChannelGroup.count).to eq(17)
    end
  end
end
