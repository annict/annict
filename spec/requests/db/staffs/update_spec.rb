# frozen_string_literal: true

describe "PATCH /db/staffs/:id", type: :request do
  context "user does not sign in" do
    let!(:person) { create(:person) }
    let!(:staff) { create(:staff) }
    let!(:old_staff) { staff.attributes }
    let!(:staff_params) do
      {
        resource_id: person.id
      }
    end

    it "user can not access this page" do
      patch "/db/staffs/#{staff.id}", params: {staff: staff_params}
      staff.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(staff.resource_id).to eq(old_staff["resource_id"])
    end
  end

  context "user who is not editor signs in" do
    let!(:person) { create(:person) }
    let!(:user) { create(:registered_user) }
    let!(:staff) { create(:staff) }
    let!(:old_staff) { staff.attributes }
    let!(:staff_params) do
      {
        resource_id: person.id
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/staffs/#{staff.id}", params: {staff: staff_params}
      staff.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(staff.resource_id).to eq(old_staff["resource_id"])
    end
  end

  context "user who is editor signs in" do
    let!(:person) { create(:person) }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:staff) { create(:staff) }
    let!(:old_staff) { staff.attributes }
    let!(:staff_params) do
      {
        resource_id: person.id
      }
    end
    let!(:attr_names) { staff_params.keys }

    before do
      login_as(user, scope: :user)
    end

    it "user can update staff" do
      expect(staff.resource_id).to eq(old_staff["resource_id"])

      patch "/db/staffs/#{staff.id}", params: {staff: staff_params}
      staff.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(staff.resource_id).to eq(person.id)
    end
  end
end
