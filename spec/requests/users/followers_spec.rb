# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username/followers", type: :request do
  it "フォロワーがいないとき、アクセスできること" do
    user = create(:registered_user)
    get "/@#{user.username}/followers"

    expect(response.status).to eq(200)
    expect(response.body).to include("フォローされていません")
  end

  it "フォロワーがいるとき、フォロワーが表示されること" do
    user = create(:registered_user)
    follower = create(:registered_user)
    create(:follow, user: follower, following: user)

    get "/@#{user.username}/followers"

    expect(response.status).to eq(200)
    expect(response.body).to include(follower.profile.name)
  end

  it "存在しないユーザー名のとき、404エラーが返されること" do
    expect {
      get "/@nonexistent_user/followers"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたユーザーのとき、404エラーが返されること" do
    user = create(:registered_user)
    user.update!(deleted_at: Time.current)

    expect {
      get "/@#{user.username}/followers"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
