# typed: false
# frozen_string_literal: true

describe "DELETE /db/channels/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:channel) { Channel.first }

    it "user can not access this page" do
      delete "/db/channels/#{channel.id}/publishing"
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(channel.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:channel) { Channel.first }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/channels/#{channel.id}/publishing"
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(channel.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:channel) { Channel.first }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/channels/#{channel.id}/publishing"
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(channel.published?).to eq(true)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:channel) { Channel.first }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish channel" do
      expect(channel.published?).to eq(true)

      delete "/db/channels/#{channel.id}/publishing"
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(channel.published?).to eq(false)
    end
  end
end
