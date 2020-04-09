# frozen_string_literal: true

describe "DELETE /db/slots/:id", type: :request do
  context "user does not sign in" do
    let!(:slot) { create(:slot, :not_deleted) }

    it "user can not access this page" do
      delete "/db/slots/#{slot.id}"
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(slot.deleted?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:slot) { create(:slot, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/slots/#{slot.id}"
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(slot.deleted?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:slot) { create(:slot, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/slots/#{slot.id}"
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(slot.deleted?).to eq(false)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:slot) { create(:slot, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can delete slot softly" do
      expect(slot.deleted?).to eq(false)

      delete "/db/slots/#{slot.id}"
      slot.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("削除しました")

      expect(slot.deleted?).to eq(true)
    end
  end
end
