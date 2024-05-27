# typed: false
# frozen_string_literal: true

describe "POST /db/organizations/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:organization) { create(:organization, :unpublished) }

    it "user can not access this page" do
      post "/db/organizations/#{organization.id}/publishing"
      organization.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(organization.published?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:organization) { create(:organization, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/organizations/#{organization.id}/publishing"
      organization.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(organization.published?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:organization) { create(:organization, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can publish organization" do
      expect(organization.published?).to eq(false)

      post "/db/organizations/#{organization.id}/publishing"
      organization.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("公開しました")

      expect(organization.published?).to eq(true)
    end
  end
end
