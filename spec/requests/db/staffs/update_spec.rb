# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/staffs/:id", type: :request do
  it "ログインしていないとき、このページにアクセスできないこと" do
    person = FactoryBot.create(:person)
    staff = FactoryBot.create(:staff)
    old_staff = staff.attributes
    staff_params = {
      resource_id: person.id
    }

    patch "/db/staffs/#{staff.id}", params: {staff: staff_params}
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")

    expect(staff.resource_id).to eq(old_staff["resource_id"])
  end

  it "編集者権限を持たないユーザーでログインしているとき、アクセスできないこと" do
    person = FactoryBot.create(:person)
    user = FactoryBot.create(:registered_user)
    staff = FactoryBot.create(:staff)
    old_staff = staff.attributes
    staff_params = {
      resource_id: person.id
    }

    login_as(user, scope: :user)

    patch "/db/staffs/#{staff.id}", params: {staff: staff_params}
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(staff.resource_id).to eq(old_staff["resource_id"])
  end

  it "編集者権限を持つユーザーでログインしているとき、スタッフを更新できること" do
    person = FactoryBot.create(:person)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    staff = FactoryBot.create(:staff)
    old_staff = staff.attributes
    staff_params = {
      resource_id: person.id
    }

    login_as(user, scope: :user)

    expect(staff.resource_id).to eq(old_staff["resource_id"])

    patch "/db/staffs/#{staff.id}", params: {staff: staff_params}
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")

    expect(staff.resource_id).to eq(person.id)
  end
end
