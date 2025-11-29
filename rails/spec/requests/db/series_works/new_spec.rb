# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/series/:series_id/series_works/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    series = FactoryBot.create(:series)

    get "/db/series/#{series.id}/series_works/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限のないユーザーがログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user)
    series = FactoryBot.create(:series)

    login_as(user, scope: :user)

    get "/db/series/#{series.id}/series_works/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限のあるユーザーがログインしているとき、ページが正常に表示されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    series = FactoryBot.create(:series)

    login_as(user, scope: :user)

    get "/db/series/#{series.id}/series_works/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("シリーズ作品登録")
  end
end
