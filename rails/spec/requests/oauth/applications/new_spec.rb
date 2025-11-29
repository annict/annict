# typed: false
# frozen_string_literal: true

RSpec.describe "GET /oauth/applications/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    get "/oauth/applications/new"

    expect(response.status).to eq(302)
    expect(response.location).to start_with("http://api.annict.test:3000/sign_in")
  end

  it "管理者でないユーザーがアクセスしたとき、新規作成フォームが表示されること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    get "/oauth/applications/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("form")
  end

  it "管理者ユーザーがアクセスしたとき、新規作成フォームが表示されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    get "/oauth/applications/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("form")
  end
end
