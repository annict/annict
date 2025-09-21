# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username/collections", type: :request do
  it "ログインしているときコレクションが存在しないとき、ページが正常に表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/@#{user.username}/collections"

    expect(response.status).to eq(200)
  end

  it "ログインしているときコレクションが存在するとき、コレクションが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    create(:collection, user: user, name: "テストコレクション")

    get "/@#{user.username}/collections"

    expect(response.status).to eq(200)
    expect(response.body).to include("テストコレクション")
  end

  it "ログインしていないときコレクションが存在しないとき、ページが正常に表示されること" do
    user = create(:registered_user)

    get "/@#{user.username}/collections"

    expect(response.status).to eq(200)
  end

  it "ログインしていないときコレクションが存在するとき、コレクションが表示されること" do
    user = create(:registered_user)
    create(:collection, user: user, name: "テストコレクション")

    get "/@#{user.username}/collections"

    expect(response.status).to eq(200)
    expect(response.body).to include("テストコレクション")
  end

  it "存在しないユーザー名でアクセスしたとき、404エラーが返されること" do
    expect {
      get "/@nonexistent_user/collections"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたユーザーにアクセスしたとき、404エラーが返されること" do
    user = create(:registered_user)
    user.update!(deleted_at: Time.current)

    expect {
      get "/@#{user.username}/collections"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたコレクションが表示されないこと" do
    user = create(:registered_user)
    create(:collection, user: user, name: "表示されるコレクション")
    create(:collection, user: user, name: "削除されたコレクション", deleted_at: Time.current)

    get "/@#{user.username}/collections"

    expect(response.status).to eq(200)
    expect(response.body).to include("表示されるコレクション")
    expect(response.body).not_to include("削除されたコレクション")
  end

  it "複数のコレクションが作成日時の降順で表示されること" do
    user = create(:registered_user)

    create(:collection, user: user, name: "古いコレクション", created_at: 2.days.ago)
    create(:collection, user: user, name: "新しいコレクション", created_at: 1.day.ago)

    get "/@#{user.username}/collections"

    expect(response.status).to eq(200)
    expect(response.body).to include("古いコレクション")
    expect(response.body).to include("新しいコレクション")

    new_index = response.body.index("新しいコレクション")
    old_index = response.body.index("古いコレクション")
    expect(new_index).to be < old_index
  end
end
