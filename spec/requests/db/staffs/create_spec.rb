# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/works/:work_id/staffs", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    person = create(:person)
    work = create(:work)
    form_params = {
      rows: "監督,#{person.id}"
    }

    post "/db/works/#{work.id}/staffs", params: {deprecated_db_staff_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Staff.all.size).to eq(0)
  end

  it "エディターロールを持たないユーザーがログインしているとき、アクセスできないこと" do
    person = create(:person)
    work = create(:work)
    user = create(:registered_user)
    form_params = {
      rows: "監督,#{person.id}"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/staffs", params: {deprecated_db_staff_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Staff.all.size).to eq(0)
  end

  it "エディターロールを持つユーザーがログインしているとき、スタッフを登録できること" do
    person = create(:person)
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "監督,#{person.id}"
    }

    login_as(user, scope: :user)

    expect(Staff.all.size).to eq(0)

    post "/db/works/#{work.id}/staffs", params: {deprecated_db_staff_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Staff.all.size).to eq(1)
    staff = Staff.last
    expect(staff.resource_id).to eq(person.id)
  end
end
