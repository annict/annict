# typed: false
# frozen_string_literal: true

describe "DELETE /db/organizations/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:organization) { create(:organization, :published) }

    it "user can not access this page" do
      delete "/db/organizations/#{organization.id}/publishing"
      organization.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(organization.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:organization) { create(:organization, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/organizations/#{organization.id}/publishing"
      organization.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(organization.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:organization) { create(:organization, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish organization" do
      expect(organization.published?).to eq(true)

      delete "/db/organizations/#{organization.id}/publishing"
      organization.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(organization.published?).to eq(false)
    end
  end
end
