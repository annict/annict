# typed: false
# frozen_string_literal: true

describe "PATCH /db/slots/:id", type: :request do
  context "user does not sign in" do
    let!(:channel) { Channel.first }
    let!(:slot) { create(:slot) }
    let!(:old_slot) { slot.attributes }
    let!(:slot_params) do
      {
        channel_id: channel.id
      }
    end

    it "user can not access this page" do
      patch "/db/slots/#{slot.id}", params: {slot: slot_params}
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(slot.channel_id).to eq(old_slot["channel_id"])
    end
  end

  context "user who is not editor signs in" do
    let!(:channel) { Channel.first }
    let!(:user) { create(:registered_user) }
    let!(:slot) { create(:slot) }
    let!(:old_slot) { slot.attributes }
    let!(:slot_params) do
      {
        channel_id: channel.id
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/slots/#{slot.id}", params: {slot: slot_params}
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(slot.channel_id).to eq(old_slot["channel_id"])
    end
  end

  context "user who is editor signs in" do
    let!(:channel) { Channel.first }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:slot) { create(:slot) }
    let!(:old_slot) { slot.attributes }
    let!(:slot_params) do
      {
        channel_id: channel.id
      }
    end
    let!(:attr_names) { slot_params.keys }

    before do
      login_as(user, scope: :user)
    end

    it "user can update slot" do
      expect(slot.channel_id).to eq(old_slot["channel_id"])

      patch "/db/slots/#{slot.id}", params: {slot: slot_params}
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(slot.channel_id).to eq(channel.id)
    end
  end
end
