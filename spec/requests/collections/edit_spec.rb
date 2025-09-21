# typed: false
# frozen_string_literal: true

RSpec.describe "GET /collections/:collection_id/edit", type: :request do
  it "ログインしているとき、自分のコレクションの編集ページが正常に表示されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, name: "テストコレクション")
    login_as(user, scope: :user)

    get "/collections/#{collection.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include("テストコレクション")
  end

  it "ログインしていないとき、ログインページにリダイレクトされること" do
    collection = create(:collection)

    get "/collections/#{collection.id}/edit"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "他のユーザーのコレクションを編集しようとしたとき、404エラーが返されること" do
    user1 = create(:registered_user)
    user2 = create(:registered_user)
    collection = create(:collection, user: user1)
    login_as(user2, scope: :user)

    expect {
      get "/collections/#{collection.id}/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないコレクションを編集しようとしたとき、404エラーが返されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/collections/99999999/edit"

    expect(response.status).to eq(404)
  end

  it "削除されたコレクションを編集しようとしたとき、404エラーが返されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user)
    collection.update!(deleted_at: Time.current)
    login_as(user, scope: :user)

    expect {
      get "/collections/#{collection.id}/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
