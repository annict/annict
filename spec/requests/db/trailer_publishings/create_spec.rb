# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/trailers/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    trailer = create(:trailer, :unpublished)

    post "/db/trailers/#{trailer.id}/publishing"
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(trailer.published?).to eq(false)
  end

  it "エディター権限のないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    trailer = create(:trailer, :unpublished)
    login_as(user, scope: :user)

    post "/db/trailers/#{trailer.id}/publishing"
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(trailer.published?).to eq(false)
  end

  it "エディター権限のあるユーザーがログインしているとき、トレイラーを公開できること" do
    user = create(:registered_user, :with_editor_role)
    trailer = create(:trailer, :unpublished)
    login_as(user, scope: :user)

    expect(trailer.published?).to eq(false)

    post "/db/trailers/#{trailer.id}/publishing"
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(trailer.published?).to eq(true)
  end

  it "管理者権限のあるユーザーがログインしているとき、トレイラーを公開できること" do
    user = create(:registered_user, :with_admin_role)
    trailer = create(:trailer, :unpublished)
    login_as(user, scope: :user)

    expect(trailer.published?).to eq(false)

    post "/db/trailers/#{trailer.id}/publishing"
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(trailer.published?).to eq(true)
  end

  it "すでに公開済みのトレイラーを公開しようとしたとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    trailer = create(:trailer, :published)
    login_as(user, scope: :user)

    expect(trailer.published?).to eq(true)

    post "/db/trailers/#{trailer.id}/publishing"

    expect(response.status).to eq(404)
  end

  it "存在しないトレイラーIDを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    post "/db/trailers/999999/publishing"

    expect(response.status).to eq(404)
  end

  it "削除済みのトレイラーを公開しようとしたとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    trailer = create(:trailer, :unpublished, deleted_at: Time.zone.now)
    login_as(user, scope: :user)

    post "/db/trailers/#{trailer.id}/publishing"

    expect(response.status).to eq(404)
  end
end
