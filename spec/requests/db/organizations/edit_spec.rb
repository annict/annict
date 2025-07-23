# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/organizations/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    organization = FactoryBot.create(:organization)

    get "/db/organizations/#{organization.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限を持たないユーザーでログインしているとき、アクセスが拒否されること" do
    user = FactoryBot.create(:registered_user)
    organization = FactoryBot.create(:organization)
    login_as(user, scope: :user)

    get "/db/organizations/#{organization.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限を持つユーザーでログインしているとき、編集フォームが表示されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    organization = FactoryBot.create(:organization)
    login_as(user, scope: :user)

    get "/db/organizations/#{organization.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(organization.name)
  end

  it "管理者権限を持つユーザーでログインしているとき、編集フォームが表示されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    organization = FactoryBot.create(:organization)
    login_as(user, scope: :user)

    get "/db/organizations/#{organization.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(organization.name)
  end

  it "削除された組織を編集しようとしたとき、404エラーが返されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    organization = FactoryBot.create(:organization, :deleted)
    login_as(user, scope: :user)

    expect {
      get "/db/organizations/#{organization.id}/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しない組織を編集しようとしたとき、404エラーが返されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect {
      get "/db/organizations/999999999/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
