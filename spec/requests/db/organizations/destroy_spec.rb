# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/organizations/:id", type: :request do
  it "ログインしていないとき、アクセスできず削除されないこと" do
    organization = create(:organization, :not_deleted)

    expect(Organization.count).to eq(1)

    delete "/db/organizations/#{organization.id}"
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Organization.count).to eq(1)
  end

  it "編集者権限を持つユーザーがサインインしているとき、アクセスできず削除されないこと" do
    user = create(:registered_user, :with_editor_role)
    organization = create(:organization, :not_deleted)
    login_as(user, scope: :user)

    expect(Organization.count).to eq(1)

    delete "/db/organizations/#{organization.id}"
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Organization.count).to eq(1)
  end

  it "編集者権限を持たないユーザーがサインインしているとき、アクセスできず削除されないこと" do
    user = create(:registered_user)
    organization = create(:organization, :not_deleted)
    login_as(user, scope: :user)

    expect(Organization.count).to eq(1)

    delete "/db/organizations/#{organization.id}"
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Organization.count).to eq(1)
  end

  it "管理者権限を持つユーザーがサインインしているとき、組織を論理削除できること" do
    user = create(:registered_user, :with_admin_role)
    organization = create(:organization, :not_deleted)
    login_as(user, scope: :user)

    expect(Organization.count).to eq(1)

    delete "/db/organizations/#{organization.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
    expect(Organization.count).to eq(0)
  end

  it "存在しない組織IDを指定したとき、404エラーが返ること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    delete "/db/organizations/non-existent-id"

    expect(response.status).to eq(404)
  end

  it "既に削除された組織を削除しようとしたとき、404エラーが返ること" do
    user = create(:registered_user, :with_admin_role)
    organization = create(:organization, :deleted)
    login_as(user, scope: :user)

    expect {
      delete "/db/organizations/#{organization.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
