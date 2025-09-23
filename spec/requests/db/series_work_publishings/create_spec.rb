# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/series_works/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    series_work = FactoryBot.create(:series_work, :unpublished)

    post "/db/series_works/#{series_work.id}/publishing"
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(series_work.published?).to eq(false)
  end

  it "エディター権限のないユーザーでログインしているとき、アクセスが拒否されること" do
    user = FactoryBot.create(:registered_user)
    series_work = FactoryBot.create(:series_work, :unpublished)
    login_as(user, scope: :user)

    post "/db/series_works/#{series_work.id}/publishing"
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(series_work.published?).to eq(false)
  end

  it "エディター権限のあるユーザーでログインしているとき、シリーズ作品を公開できること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    series_work = FactoryBot.create(:series_work, :unpublished)
    login_as(user, scope: :user)

    expect(series_work.published?).to eq(false)

    post "/db/series_works/#{series_work.id}/publishing"
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(series_work.published?).to eq(true)
  end

  it "存在しないシリーズ作品のIDが指定されたとき、404エラーが返されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    post "/db/series_works/non-existent-id/publishing"

    expect(response.status).to eq(404)
  end

  it "既に公開済みのシリーズ作品のIDが指定されたとき、404エラーが返されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    series_work = FactoryBot.create(:series_work, :published)
    login_as(user, scope: :user)

    post "/db/series_works/#{series_work.id}/publishing"

    expect(response.status).to eq(404)
  end
end
