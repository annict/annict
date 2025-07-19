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

  it "存在しないユーザーにアクセスしたとき、RecordNotFoundエラーが発生すること" do
    collection = create(:collection)

    expect {
      get "/@nonexistent_user/collections/#{collection.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたユーザーにアクセスしたとき、RecordNotFoundエラーが発生すること" do
    user = create(:registered_user)
    collection = create(:collection, user: user)
    user.update!(deleted_at: Time.current)

    expect {
      get "/@#{user.username}/collections/#{collection.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないコレクションにアクセスしたとき、RecordNotFoundエラーが発生すること" do
    user = create(:registered_user)

    expect {
      get "/@#{user.username}/collections/99999999"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたコレクションにアクセスしたとき、RecordNotFoundエラーが発生すること" do
    user = create(:registered_user)
    collection = create(:collection, user: user)
    collection.update!(deleted_at: Time.current)

    expect {
      get "/@#{user.username}/collections/#{collection.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "他のユーザーのコレクションにアクセスしたとき、RecordNotFoundエラーが発生すること" do
    user1 = create(:registered_user)
    user2 = create(:registered_user)
    collection = create(:collection, user: user1)

    expect {
      get "/@#{user2.username}/collections/#{collection.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
