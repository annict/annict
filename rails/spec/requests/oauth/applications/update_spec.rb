# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /oauth/applications/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    application = FactoryBot.create(:oauth_application)

    patch "/oauth/applications/#{application.id}", params: {
      oauth_application: {name: "Updated Name"}
    }

    expect(response.status).to eq(302)
    expect(response.location).to start_with("http://api.annict.test:3000/sign_in")
  end

  it "管理者でないユーザーが自分のアプリケーションを更新したとき、詳細ページにリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    application = FactoryBot.create(:oauth_application, owner: user)
    login_as(user, scope: :user)

    patch "/oauth/applications/#{application.id}", params: {
      oauth_application: {name: "Updated Name"}
    }

    expect(response.status).to eq(302)
    expect(response.location).to eq("http://api.annict.test:3000/oauth/applications/#{application.id}")
    expect(application.reload.name).to eq("Updated Name")
  end

  it "管理者ユーザーが自分のアプリケーションを正常に更新したとき、詳細ページにリダイレクトされること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user, name: "Original Name")
    login_as(user, scope: :user)

    patch "/oauth/applications/#{application.id}", params: {
      oauth_application: {name: "Updated Name"}
    }

    expect(response.status).to eq(302)
    expect(response.location).to eq("http://api.annict.test:3000/oauth/applications/#{application.id}")
    expect(application.reload.name).to eq("Updated Name")
  end

  it "管理者ユーザーが無効なパラメータで更新したとき、編集フォームが再表示されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user, name: "Original Name")
    login_as(user, scope: :user)

    patch "/oauth/applications/#{application.id}", params: {
      oauth_application: {name: ""}
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("form")
    expect(application.reload.name).to eq("Original Name")
  end

  it "管理者ユーザーが他のユーザーのアプリケーションを更新しようとしたとき、404エラーが発生すること" do
    user1 = FactoryBot.create(:registered_user, :with_admin_role)
    user2 = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user2)
    login_as(user1, scope: :user)

    expect do
      patch "/oauth/applications/#{application.id}", params: {
        oauth_application: {name: "Updated Name"}
      }
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "管理者ユーザーが削除済みアプリケーションを更新しようとしたとき、404エラーが発生すること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user)
    application.update!(deleted_at: Time.current)
    login_as(user, scope: :user)

    expect do
      patch "/oauth/applications/#{application.id}", params: {
        oauth_application: {name: "Updated Name"}
      }
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないアプリケーションIDで更新しようとしたとき、404エラーが発生すること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    expect do
      patch "/oauth/applications/non-existent-id", params: {
        oauth_application: {name: "Updated Name"}
      }
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "管理者ユーザーがリダイレクトURIを更新したとき、正常に更新されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user, redirect_uri: "https://example.com/callback")
    login_as(user, scope: :user)

    patch "/oauth/applications/#{application.id}", params: {
      oauth_application: {redirect_uri: "https://updated.com/callback"}
    }

    expect(response.status).to eq(302)
    expect(application.reload.redirect_uri).to eq("https://updated.com/callback")
  end

  it "管理者ユーザーがスコープを更新したとき、正常に更新されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user, scopes: "read")
    login_as(user, scope: :user)

    patch "/oauth/applications/#{application.id}", params: {
      oauth_application: {scopes: "read write"}
    }

    expect(response.status).to eq(302)
    expect(application.reload.scopes.to_s).to eq("read write")
  end
end
