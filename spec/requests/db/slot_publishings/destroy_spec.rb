# typed: false
# frozen_string_literal: true

describe "DELETE /db/slots/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:slot) { create(:slot, :published) }

    it "user can not access this page" do
      delete "/db/slots/#{slot.id}/publishing"
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(slot.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:slot) { create(:slot, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/slots/#{slot.id}/publishing"
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(slot.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:slot) { create(:slot, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish slot" do
      expect(slot.published?).to eq(true)

      delete "/db/slots/#{slot.id}/publishing"
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(slot.published?).to eq(false)
    end
  end
end
