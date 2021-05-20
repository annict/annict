# frozen_string_literal: true

describe "DELETE /db/channel_groups/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:channel_group) { ChannelGroup.first }

    it "user can not access this page" do
      delete "/db/channel_groups/#{channel_group.id}/publishing"
      channel_group.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(channel_group.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:channel_group) { ChannelGroup.first }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/channel_groups/#{channel_group.id}/publishing"
      channel_group.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(channel_group.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:channel_group) { ChannelGroup.first }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/channel_groups/#{channel_group.id}/publishing"
      channel_group.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(channel_group.published?).to eq(true)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:channel_group) { ChannelGroup.first }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish channel_group" do
      expect(channel_group.published?).to eq(true)

      delete "/db/channel_groups/#{channel_group.id}/publishing"
      channel_group.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(channel_group.published?).to eq(false)
    end
  end
end
