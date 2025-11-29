# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/people/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/db/people/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限がないユーザーとしてログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/db/people/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限があるユーザーとしてログインしているとき、人物登録ページが表示されること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/people/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("人物登録")
  end
end
