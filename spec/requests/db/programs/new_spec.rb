# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/programs/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = FactoryBot.create(:work)

    get "/db/works/#{work.id}/programs/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/programs/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限を持つユーザーがログインしているとき、ページが表示されること" do
    work = FactoryBot.create(:work)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/programs/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("放送情報登録")
  end

  it "存在しない作品IDを指定したとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect {
      get "/db/works/99999999/programs/new"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除された作品を指定したとき、404エラーになること" do
    work = FactoryBot.create(:work, deleted_at: Time.current)
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect {
      get "/db/works/#{work.id}/programs/new"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
