# typed: false
# frozen_string_literal: true

describe "POST /db/channels/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:channel) { Channel.first.tap { |c| c.unpublish } }

    it "user can not access this page" do
      post "/db/channels/#{channel.id}/publishing"
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(channel.published?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:channel) { Channel.first.tap { |c| c.unpublish } }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/channels/#{channel.id}/publishing"
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(channel.published?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:channel) { Channel.first.tap { |c| c.unpublish } }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/channels/#{channel.id}/publishing"
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(channel.published?).to eq(false)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:channel) { Channel.first.tap { |c| c.unpublish } }

    before do
      login_as(user, scope: :user)
    end

    it "user can publish channel" do
      expect(channel.published?).to eq(false)

      post "/db/channels/#{channel.id}/publishing"
      channel.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("公開しました")

      expect(channel.published?).to eq(true)
    end
  end
end
