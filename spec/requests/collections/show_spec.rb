# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username/collections/:collection_id", type: :request do
  it "ログインしているとき、コレクションが正常に表示されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, name: "テストコレクション")
    login_as(user, scope: :user)

    get "/@#{user.username}/collections/#{collection.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include("テストコレクション")
  end

  it "ログインしていないとき、コレクションが正常に表示されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, name: "テストコレクション")

    get "/@#{user.username}/collections/#{collection.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include("テストコレクション")
  end

  it "存在しないユーザーにアクセスしたとき、404エラーが返されること" do
    collection = create(:collection)

    get "/@nonexistent_user/collections/#{collection.id}"

    expect(response.status).to eq(404)
  end

  it "削除されたユーザーにアクセスしたとき、404エラーが返されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user)
    user.update!(deleted_at: Time.current)

    get "/@#{user.username}/collections/#{collection.id}"

    expect(response.status).to eq(404)
  end

  it "存在しないコレクションにアクセスしたとき、404エラーが返されること" do
    user = create(:registered_user)

    get "/@#{user.username}/collections/99999999"

    expect(response.status).to eq(404)
  end

  it "削除されたコレクションにアクセスしたとき、404エラーが返されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user)
    collection.update!(deleted_at: Time.current)

    get "/@#{user.username}/collections/#{collection.id}"

    expect(response.status).to eq(404)
  end

  it "他のユーザーのコレクションにアクセスしたとき、404エラーが返されること" do
    user1 = create(:registered_user)
    user2 = create(:registered_user)
    collection = create(:collection, user: user1)

    get "/@#{user2.username}/collections/#{collection.id}"

    expect(response.status).to eq(404)
  end
end
