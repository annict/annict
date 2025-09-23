# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/casts/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = FactoryBot.create(:work)

    get "/db/works/#{work.id}/casts/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/casts/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限を持つユーザーがログインしているとき、ページが表示されること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/casts/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("キャスト登録")
  end

  it "管理者権限を持つユーザーがログインしているとき、ページが表示されること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/casts/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("キャスト登録")
  end

  it "存在しない作品IDが指定されたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/999999/casts/new"

    expect(response.status).to eq(404)
  end

  it "削除済みの作品に対してアクセスしたとき、404エラーになること" do
    work = FactoryBot.create(:work, deleted_at: Time.current)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/casts/new"

    expect(response.status).to eq(404)
  end
end
