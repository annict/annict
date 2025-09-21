# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username/following", type: :request do
  it "フォロイーがいないとき、アクセスできること" do
    user = create(:registered_user)
    get "/@#{user.username}/following"

    expect(response.status).to eq(200)
    expect(response.body).to include("フォローしていません")
  end

  it "フォロイーがいるとき、フォロイーが表示されること" do
    user = create(:registered_user)
    followee = create(:registered_user)
    create(:follow, user: user, following: followee)

    get "/@#{user.username}/following"

    expect(response.status).to eq(200)
    expect(response.body).to include(followee.profile.name)
  end

  it "存在しないユーザー名のとき、404エラーが返されること" do
    get "/@nonexistent_user/following"

    expect(response.status).to eq(404)
  end

  it "削除されたユーザーのとき、404エラーが返されること" do
    user = create(:registered_user)
    user.update!(deleted_at: Time.current)

    expect {
      get "/@#{user.username}/following"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
