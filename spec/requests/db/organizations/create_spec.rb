# frozen_string_literal: true

describe "POST /db/organizations", type: :request do
  context "user does not sign in" do
    let!(:organization_params) do
      {
        rows: "御三家,ごさんけ"
      }
    end

    it "user can not access this page" do
      post "/db/organizations", params: {db_organization_rows_form: organization_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Organization.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:organization_params) do
      {
        rows: "御三家,ごさんけ"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/organizations", params: {db_organization_rows_form: organization_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Organization.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:organization_params) do
      {
        rows: "御三家,ごさんけ"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create organization" do
      expect(Organization.all.size).to eq(0)

      post "/db/organizations", params: {db_organization_rows_form: organization_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Organization.all.size).to eq(1)
      organization = Organization.first

      expect(organization.name).to eq("御三家")
      expect(organization.name_kana).to eq("ごさんけ")
    end
  end
end
