# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/works/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = create(:work, :unpublished)

    post "/db/works/#{work.id}/publishing"
    work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(work.published?).to eq(false)
  end

  it "エディター権限を持たないユーザーの場合、アクセスできないこと" do
    user = create(:registered_user)
    work = create(:work, :unpublished)
    login_as(user, scope: :user)

    post "/db/works/#{work.id}/publishing"
    work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(work.published?).to eq(false)
  end

  it "エディター権限を持つユーザーの場合、作品を公開できること" do
    user = create(:registered_user, :with_editor_role)
    work = create(:work, :unpublished)
    login_as(user, scope: :user)

    expect(work.published?).to eq(false)

    post "/db/works/#{work.id}/publishing"
    work.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(work.published?).to eq(true)
  end

  it "すでに公開済みの作品の場合、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    work = create(:work, :published)
    login_as(user, scope: :user)

    expect { post "/db/works/#{work.id}/publishing" }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済みの作品の場合、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    work = create(:work, :unpublished)
    work.destroy!
    login_as(user, scope: :user)

    expect { post "/db/works/#{work.id}/publishing" }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しない作品IDの場合、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    post "/db/works/invalid-id/publishing"

    expect(response).to have_http_status(:not_found)
  end
end
