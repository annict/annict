# typed: false
# frozen_string_literal: true

RSpec.describe "GET /oauth/applications", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    get "/oauth/applications"

    expect(response.status).to eq(302)
    expect(response.location).to start_with("http://api.annict.test:3000/sign_in")
  end

  it "管理者でないユーザーがアクセスしたとき、アプリケーション一覧が表示されること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    get "/oauth/applications"

    expect(response.status).to eq(200)
  end

  it "管理者ユーザーがアクセスしたとき、アプリケーション一覧が表示されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    FactoryBot.create(:oauth_application, owner: user, name: "Test App 1")
    FactoryBot.create(:oauth_application, owner: user, name: "Test App 2")
    login_as(user, scope: :user)

    get "/oauth/applications"

    expect(response.status).to eq(200)
    expect(response.body).to include("Test App 1")
    expect(response.body).to include("Test App 2")
  end

  it "管理者ユーザーが削除済みアプリケーションを持っているとき、削除済みアプリケーションは表示されないこと" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    FactoryBot.create(:oauth_application, owner: user, name: "Active App")
    deleted_app = FactoryBot.create(:oauth_application, owner: user, name: "Deleted App")
    deleted_app.update!(deleted_at: Time.current)
    login_as(user, scope: :user)

    get "/oauth/applications"

    expect(response.status).to eq(200)
    expect(response.body).to include("Active App")
    expect(response.body).not_to include("Deleted App")
  end

  it "管理者ユーザーがアプリケーションを持っていないとき、空の一覧が表示されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    get "/oauth/applications"

    expect(response.status).to eq(200)
  end

  it "管理者ユーザーが他のユーザーのアプリケーションを持っているとき、自分のアプリケーションのみ表示されること" do
    user1 = FactoryBot.create(:registered_user, :with_admin_role)
    user2 = FactoryBot.create(:registered_user, :with_admin_role)
    FactoryBot.create(:oauth_application, owner: user1, name: "My App")
    FactoryBot.create(:oauth_application, owner: user2, name: "Other User App")
    login_as(user1, scope: :user)

    get "/oauth/applications"

    expect(response.status).to eq(200)
    expect(response.body).to include("My App")
    expect(response.body).not_to include("Other User App")
  end
end
