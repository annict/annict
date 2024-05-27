# typed: false
# frozen_string_literal: true

describe "POST /db/slots/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:slot) { create(:slot, :unpublished) }

    it "user can not access this page" do
      post "/db/slots/#{slot.id}/publishing"
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(slot.published?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:slot) { create(:slot, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/slots/#{slot.id}/publishing"
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(slot.published?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:slot) { create(:slot, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can publish slot" do
      expect(slot.published?).to eq(false)

      post "/db/slots/#{slot.id}/publishing"
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("公開しました")

      expect(slot.published?).to eq(true)
    end
  end
end
