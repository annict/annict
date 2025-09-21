# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/works/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    work = create(:work, :published)

    delete "/db/works/#{work.id}/publishing"
    work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(work.published?).to eq(true)
  end

  it "エディター権限がないユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    work = create(:work, :published)
    login_as(user, scope: :user)

    delete "/db/works/#{work.id}/publishing"
    work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(work.published?).to eq(true)
  end

  it "エディター権限があるユーザーでログインしているとき、作品を非公開にできること" do
    user = create(:registered_user, :with_editor_role)
    work = create(:work, :published)
    login_as(user, scope: :user)

    expect(work.published?).to eq(true)

    delete "/db/works/#{work.id}/publishing"
    work.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(work.published?).to eq(false)
  end

  it "エディター権限があるユーザーでログインしているとき、削除済みの作品は見つからないこと" do
    user = create(:registered_user, :with_editor_role)
    work = create(:work, :published, deleted_at: Time.current)
    login_as(user, scope: :user)

    delete "/db/works/#{work.id}/publishing"

    expect(response).to have_http_status(404)
  end

  it "エディター権限があるユーザーでログインしているとき、未公開の作品は見つからないこと" do
    user = create(:registered_user, :with_editor_role)
    work = create(:work, :unpublished)
    login_as(user, scope: :user)

    delete "/db/works/#{work.id}/publishing"

    expect(response).to have_http_status(404)
  end

  it "エディター権限があるユーザーでログインしているとき、存在しない作品IDの場合は見つからないこと" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    delete "/db/works/nonexistent-id/publishing"

    expect(response).to have_http_status(404)
  end
end
