# typed: false
# frozen_string_literal: true

describe "GET /db/works/:work_id/staffs", type: :request do
  context "user does not sign in" do
    let!(:staff) { create(:staff) }

    it "responses staff list" do
      get "/db/works/#{staff.work_id}/staffs"

      expect(response.status).to eq(200)
      expect(response.body).to include(staff.resource.name)
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }
    let!(:staff) { create(:staff) }

    before do
      login_as(user, scope: :user)
    end

    it "responses work list" do
      get "/db/works/#{staff.work_id}/staffs"

      expect(response.status).to eq(200)
      expect(response.body).to include(staff.resource.name)
    end
  end
end
