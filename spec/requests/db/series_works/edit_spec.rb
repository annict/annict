# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/series_works/:id/edit", type: :request do
  it "ログインしていないとき、このページにアクセスできないこと" do
    series_work = create(:series_work)

    get "/db/series_works/#{series_work.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限のないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    series_work = create(:series_work)
    login_as(user, scope: :user)

    get "/db/series_works/#{series_work.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限のあるユーザーがログインしているとき、シリーズ作品フォームが表示されること" do
    user = create(:registered_user, :with_editor_role)
    series_work = create(:series_work)
    work = series_work.work
    login_as(user, scope: :user)

    get "/db/series_works/#{series_work.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "エディター権限のあるユーザーがログインしているとき、存在しないシリーズ作品IDでアクセスするとNotFoundエラーが発生すること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/series_works/00000000-0000-0000-0000-000000000000/edit"

    expect(response.status).to eq(404)
  end
end
