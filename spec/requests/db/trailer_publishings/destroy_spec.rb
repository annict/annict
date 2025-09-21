# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/trailers/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    trailer = FactoryBot.create(:trailer, :published)

    delete "/db/trailers/#{trailer.id}/publishing"
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(trailer.published?).to eq(true)
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスが拒否されること" do
    user = FactoryBot.create(:registered_user)
    trailer = FactoryBot.create(:trailer, :published)
    login_as(user, scope: :user)

    delete "/db/trailers/#{trailer.id}/publishing"
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(trailer.published?).to eq(true)
  end

  it "エディター権限を持つユーザーがログインしているとき、トレーラーを非公開にできること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    trailer = FactoryBot.create(:trailer, :published)
    login_as(user, scope: :user)

    expect(trailer.published?).to eq(true)

    delete "/db/trailers/#{trailer.id}/publishing"
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(trailer.published?).to eq(false)
  end

  it "存在しないトレーラーIDを指定したとき、404エラーが返されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    delete "/db/trailers/nonexistent-id/publishing"

    expect(response).to have_http_status(:not_found)
  end

  it "すでに非公開のトレーラーを非公開にしようとしたとき、404エラーが返されること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    trailer = FactoryBot.create(:trailer, :unpublished)
    login_as(user, scope: :user)

    expect(trailer.published?).to eq(false)

    delete "/db/trailers/#{trailer.id}/publishing"

    expect(response).to have_http_status(:not_found)
  end
end
