# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/staffs/new", type: :request do
  it "ログインしていない場合、ログインページにリダイレクトされること" do
    work = FactoryBot.create(:work)

    get "/db/works/#{work.id}/staffs/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限がないユーザーでログインしている場合、アクセスできないこと" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/staffs/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限があるユーザーでログインしている場合、ページが表示されること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/staffs/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("スタッフ登録")
  end
end
