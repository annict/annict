# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/trailers/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    trailer = FactoryBot.create(:trailer)

    get "/db/trailers/#{trailer.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者でないユーザーでログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user)
    trailer = FactoryBot.create(:trailer)
    login_as(user, scope: :user)

    get "/db/trailers/#{trailer.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者ユーザーでログインしているとき、トレイラー編集フォームが表示されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    trailer = FactoryBot.create(:trailer)
    login_as(user, scope: :user)

    get "/db/trailers/#{trailer.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(trailer.title)
  end

  it "管理者ユーザーでログインしているとき、トレイラー編集フォームが表示されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    trailer = FactoryBot.create(:trailer)
    login_as(user, scope: :user)

    get "/db/trailers/#{trailer.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(trailer.title)
  end

  it "削除されたトレイラーの場合、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    trailer = FactoryBot.create(:trailer)
    trailer.destroy!
    login_as(user, scope: :user)

    expect {
      get "/db/trailers/#{trailer.id}/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないトレイラーIDの場合、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect {
      get "/db/trailers/non-existent-id/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
