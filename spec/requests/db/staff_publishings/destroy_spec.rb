# typed: false
# frozen_string_literal: true

describe "DELETE /db/staffs/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:staff) { create(:staff, :published) }

    it "user can not access this page" do
      delete "/db/staffs/#{staff.id}/publishing"
      staff.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(staff.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:staff) { create(:staff, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/staffs/#{staff.id}/publishing"
      staff.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(staff.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:staff) { create(:staff, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish staff" do
      expect(staff.published?).to eq(true)

      delete "/db/staffs/#{staff.id}/publishing"
      staff.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(staff.published?).to eq(false)
    end
  end
end
