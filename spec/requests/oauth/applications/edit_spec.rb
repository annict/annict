# typed: false
# frozen_string_literal: true

RSpec.describe "GET /oauth/applications/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    application = FactoryBot.create(:oauth_application)

    get "/oauth/applications/#{application.id}/edit"

    expect(response.status).to eq(302)
    expect(response.location).to start_with("http://api.annict.test:3000/sign_in")
  end

  it "管理者でないユーザーがアクセスしたとき、編集フォームが表示されること" do
    user = FactoryBot.create(:registered_user)
    application = FactoryBot.create(:oauth_application, owner: user)
    login_as(user, scope: :user)

    get "/oauth/applications/#{application.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include("form")
  end

  it "管理者ユーザーが自分のアプリケーションにアクセスしたとき、編集フォームが表示されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user, name: "My Test App")
    login_as(user, scope: :user)

    get "/oauth/applications/#{application.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include("form")
    expect(response.body).to include("My Test App")
  end

  it "管理者ユーザーが他のユーザーのアプリケーションにアクセスしたとき、404エラーが返されること" do
    user1 = FactoryBot.create(:registered_user, :with_admin_role)
    user2 = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user2)
    login_as(user1, scope: :user)

    get "/oauth/applications/#{application.id}/edit"

    expect(response).to have_http_status(:not_found)
  end

  it "管理者ユーザーが削除済みアプリケーションにアクセスしたとき、404エラーが返されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user)
    application.update!(deleted_at: Time.current)
    login_as(user, scope: :user)

    get "/oauth/applications/#{application.id}/edit"

    expect(response).to have_http_status(:not_found)
  end

  it "存在しないアプリケーションIDでアクセスしたとき、404エラーが返されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    get "/oauth/applications/non-existent-id/edit"

    expect(response).to have_http_status(:not_found)
  end
end
