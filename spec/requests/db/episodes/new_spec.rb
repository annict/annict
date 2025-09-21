# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/episodes/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = create(:work)

    get "/db/works/#{work.id}/episodes/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディターではないユーザーでログインしているとき、アクセスできないこと" do
    work = create(:work)
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/episodes/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディターのユーザーでログインしているとき、エピソード登録ページが表示されること" do
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/episodes/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("エピソード登録")
  end

  it "存在しない作品のエピソード登録ページにアクセスしたとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/99999/episodes/new"

    expect(response.status).to eq(404)
  end

  it "削除された作品のエピソード登録ページにアクセスしたとき、404エラーになること" do
    work = create(:work, deleted_at: Time.current)
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect {
      get "/db/works/#{work.id}/episodes/new"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
