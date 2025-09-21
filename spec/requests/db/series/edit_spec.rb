# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/series/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    series = FactoryBot.create(:series)

    get "/db/series/#{series.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集権限のないユーザーがアクセスしたとき、アクセス拒否されること" do
    user = FactoryBot.create(:registered_user)
    series = FactoryBot.create(:series)
    login_as(user, scope: :user)

    get "/db/series/#{series.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者がアクセスしたとき、シリーズ編集フォームが表示されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    series = FactoryBot.create(:series)
    login_as(user, scope: :user)

    get "/db/series/#{series.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
  end

  it "管理者がアクセスしたとき、シリーズ編集フォームが表示されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    series = FactoryBot.create(:series)
    login_as(user, scope: :user)

    get "/db/series/#{series.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
  end

  it "削除されたシリーズにアクセスしたとき、404エラーが発生すること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    series = FactoryBot.create(:series, deleted_at: Time.current)
    login_as(user, scope: :user)

    get "/db/series/#{series.id}/edit"

    expect(response.status).to eq(404)
  end

  it "存在しないシリーズにアクセスしたとき、404エラーが発生すること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/series/999999/edit"

    expect(response.status).to eq(404)
  end
end
