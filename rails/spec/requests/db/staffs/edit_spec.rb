# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/staffs/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    staff = create(:staff)

    get "/db/staffs/#{staff.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限がないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    staff = create(:staff)
    login_as(user, scope: :user)

    get "/db/staffs/#{staff.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限があるユーザーがログインしているとき、スタッフ編集フォームが表示されること" do
    user = create(:registered_user, :with_editor_role)
    staff = create(:staff)
    login_as(user, scope: :user)

    get "/db/staffs/#{staff.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(staff.resource.name)
  end
end
