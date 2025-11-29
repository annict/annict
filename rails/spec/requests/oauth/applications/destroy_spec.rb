# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /oauth/applications/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    application = FactoryBot.create(:oauth_application)

    delete "/oauth/applications/#{application.id}"

    expect(response.status).to eq(302)
    expect(response.location).to start_with("http://api.annict.test:3000/sign_in")
  end

  it "管理者でないユーザーが自分のアプリケーションを削除したとき、アプリケーション一覧にリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    application = FactoryBot.create(:oauth_application, owner: user)
    application_id = application.id
    login_as(user, scope: :user)

    expect {
      delete "/oauth/applications/#{application_id}"
    }.to change(Oauth::Application, :count).by(-1)

    expect(response.status).to eq(302)
    expect(response.location).to eq("http://api.annict.test:3000/oauth/applications")
  end

  it "管理者ユーザーが自分のアプリケーションを削除したとき、アプリケーション一覧にリダイレクトされること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user)
    application_id = application.id
    login_as(user, scope: :user)

    expect {
      delete "/oauth/applications/#{application_id}"
    }.to change(Oauth::Application, :count).by(-1)

    expect(response.status).to eq(302)
    expect(response.location).to eq("http://api.annict.test:3000/oauth/applications")
  end

  it "管理者ユーザーが他のユーザーのアプリケーションを削除しようとしたとき、404エラーが発生すること" do
    user1 = FactoryBot.create(:registered_user, :with_admin_role)
    user2 = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user2)
    login_as(user1, scope: :user)

    expect do
      delete "/oauth/applications/#{application.id}"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "管理者ユーザーが削除済みアプリケーションを削除しようとしたとき、404エラーが発生すること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user)
    application.update!(deleted_at: Time.current)
    login_as(user, scope: :user)

    expect do
      delete "/oauth/applications/#{application.id}"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないアプリケーションIDで削除しようとしたとき、404エラーが発生すること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    expect do
      delete "/oauth/applications/non-existent-id"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "管理者でないユーザーが他のユーザーのアプリケーションを削除しようとしたとき、404エラーが発生すること" do
    user1 = FactoryBot.create(:registered_user)
    user2 = FactoryBot.create(:registered_user)
    application = FactoryBot.create(:oauth_application, owner: user2)
    login_as(user1, scope: :user)

    expect do
      delete "/oauth/applications/#{application.id}"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "アプリケーションが削除されたとき、関連するアクセストークンも削除されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    application = FactoryBot.create(:oauth_application, owner: user)
    access_token = FactoryBot.create(:oauth_access_token, application:)
    application_id = application.id
    access_token_id = access_token.id
    login_as(user, scope: :user)

    expect {
      delete "/oauth/applications/#{application_id}"
    }.to change(Oauth::Application, :count).by(-1)
      .and change(Oauth::AccessToken, :count).by(-1)

    expect(response.status).to eq(302)
    expect(Oauth::Application.find_by(id: application_id)).to be_nil
    expect(Oauth::AccessToken.find_by(id: access_token_id)).to be_nil
  end
end
