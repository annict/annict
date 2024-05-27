# typed: false
# frozen_string_literal: true

describe "PATCH /db/organizations/:id", type: :request do
  context "user does not sign in" do
    let!(:organization) { create(:organization) }
    let!(:old_organization) { organization.attributes }
    let!(:organization_params) do
      {
        name: "御三家"
      }
    end

    it "user can not access this page" do
      patch "/db/organizations/#{organization.id}", params: {organization: organization_params}
      organization.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(organization.name).to eq(old_organization["name"])
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:organization) { create(:organization) }
    let!(:old_organization) { organization.attributes }
    let!(:organization_params) do
      {
        name: "御三家"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/organizations/#{organization.id}", params: {organization: organization_params}
      organization.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(organization.name).to eq(old_organization["name"])
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:organization) { create(:organization) }
    let!(:old_organization) { organization.attributes }
    let!(:organization_params) do
      {
        name: "御三家"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can update organization" do
      expect(organization.name).to eq(old_organization["name"])

      patch "/db/organizations/#{organization.id}", params: {organization: organization_params}
      organization.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(organization.name).to eq("御三家")
    end
  end
end
