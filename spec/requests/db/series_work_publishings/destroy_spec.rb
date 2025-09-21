# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/series_works/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされ、シリーズ作品は公開されたままであること" do
    series_work = create(:series_work, :published)

    delete "/db/series_works/#{series_work.id}/publishing"
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(series_work.published?).to eq(true)
  end

  it "編集者でないユーザーがログインしているとき、アクセスできず、シリーズ作品は公開されたままであること" do
    user = create(:registered_user)
    series_work = create(:series_work, :published)

    login_as(user, scope: :user)
    delete "/db/series_works/#{series_work.id}/publishing"
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(series_work.published?).to eq(true)
  end

  it "編集者がログインしているとき、シリーズ作品を非公開にできること" do
    user = create(:registered_user, :with_editor_role)
    series_work = create(:series_work, :published)

    login_as(user, scope: :user)

    expect(series_work.published?).to eq(true)

    delete "/db/series_works/#{series_work.id}/publishing"
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(series_work.published?).to eq(false)
  end

  it "存在しないシリーズ作品IDを指定したとき、エラーが発生すること" do
    user = create(:registered_user, :with_editor_role)

    login_as(user, scope: :user)

    delete "/db/series_works/non-existent-id/publishing"

    expect(response).to have_http_status(:not_found)
  end
end
