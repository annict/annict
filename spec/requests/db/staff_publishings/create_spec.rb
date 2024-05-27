# typed: false
# frozen_string_literal: true

describe "POST /db/staffs/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:staff) { create(:staff, :unpublished) }

    it "user can not access this page" do
      post "/db/staffs/#{staff.id}/publishing"
      staff.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(staff.published?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:staff) { create(:staff, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/staffs/#{staff.id}/publishing"
      staff.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(staff.published?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:staff) { create(:staff, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can publish staff" do
      expect(staff.published?).to eq(false)

      post "/db/staffs/#{staff.id}/publishing"
      staff.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("公開しました")

      expect(staff.published?).to eq(true)
    end
  end
end
