# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/follow", type: :request do
  it "未ログイン時は302ステータスを返すこと" do
    user = create(:user, :with_email_notification)
    post "/api/internal/follow", params: {user_id: user.id}

    expect(response.status).to eq(302)
  end

  it "ログイン時はユーザーをフォローし201ステータスを返すこと" do
    user = create(:user, :with_email_notification)
    follower = create(:user, :with_email_notification)

    expect(follower.following?(user)).to be(false)

    login_as(follower, scope: :user)
    post "/api/internal/follow", params: {user_id: user.id}

    expect(response.status).to eq(201)
    expect(follower.following?(user)).to be(true)
  end

  it "既にフォロー済みの場合でも201ステータスを返すこと" do
    user = create(:user, :with_email_notification)
    follower = create(:user, :with_email_notification)
    follower.follow(user)

    expect(follower.following?(user)).to be(true)

    login_as(follower, scope: :user)
    post "/api/internal/follow", params: {user_id: user.id}

    expect(response.status).to eq(201)
    expect(follower.following?(user)).to be(true)
  end

  it "存在しないユーザーIDを指定した場合は404エラーが返されること" do
    follower = create(:user, :with_email_notification)

    login_as(follower, scope: :user)
    expect {
      post "/api/internal/follow", params: {user_id: "nonexistent"}
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "user_idパラメータが不正な場合は404エラーが返されること" do
    follower = create(:user, :with_email_notification)

    login_as(follower, scope: :user)
    expect {
      post "/api/internal/follow", params: {}
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
