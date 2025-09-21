# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = FactoryBot.create(:work)

    get "/db/works/#{work.id}/edit"

    expect(response).to redirect_to("/sign_in")
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/edit"

    expect(response).to redirect_to(db_root_path)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限を持つユーザーがログインしているとき、編集フォームが表示されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/edit"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(work.title)
  end

  it "管理者権限を持つユーザーがログインしているとき、編集フォームが表示されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/edit"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(work.title)
  end

  it "削除済みの作品に対してアクセスしようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    work = FactoryBot.create(:work, deleted_at: Time.current)
    login_as(user, scope: :user)

    get "/db/works/#{work.id

    expect(response.status).to eq(404)
  end

  it "存在しない作品IDでアクセスしようとしたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/non-existent-id/edit"

    expect(response.status).to eq(404)
  end
end
