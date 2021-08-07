# frozen_string_literal: true

describe "POST /db/works/:work_id/staffs", type: :request do
  context "user does not sign in" do
    let!(:person) { create(:person) }
    let!(:work) { create(:anime) }
    let!(:form_params) do
      {
        rows: "監督,#{person.id}"
      }
    end

    it "user can not access this page" do
      post "/db/works/#{work.id}/staffs", params: {db_staff_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Staff.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:person) { create(:person) }
    let!(:work) { create(:anime) }
    let!(:user) { create(:registered_user) }
    let!(:form_params) do
      {
        rows: "監督,#{person.id}"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/works/#{work.id}/staffs", params: {db_staff_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Staff.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:person) { create(:person) }
    let!(:work) { create(:anime) }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:form_params) do
      {
        rows: "監督,#{person.id}"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create staff" do
      expect(Staff.all.size).to eq(0)

      post "/db/works/#{work.id}/staffs", params: {db_staff_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Staff.all.size).to eq(1)
      staff = Staff.last

      expect(staff.resource_id).to eq(person.id)
    end
  end
end
