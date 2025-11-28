# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/organizations/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/db/organizations/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限がないユーザーがログインしているとき、アクセスを拒否すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/db/organizations/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限があるユーザーがログインしているとき、団体登録ページが表示されること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/organizations/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("団体登録")
  end

  it "管理者権限があるユーザーがログインしているとき、団体登録ページが表示されること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    get "/db/organizations/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("団体登録")
  end
end
