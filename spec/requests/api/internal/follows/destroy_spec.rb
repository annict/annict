# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /api/internal/follow", type: :request do
  it "未ログイン時は302ステータスを返すこと" do
    user = create(:user)
    delete "/api/internal/follow", params: {user_id: user.id}

    expect(response.status).to eq(302)
  end

  it "ログイン時はユーザーのフォローを解除し200ステータスを返すこと" do
    user = create(:user)
    follower = create(:user)
    follower.follow(user)

    expect(follower.following?(user)).to be(true)

    login_as(follower, scope: :user)
    delete "/api/internal/follow", params: {user_id: user.id}

    expect(response.status).to eq(200)
    expect(follower.following?(user)).to be(false)
  end

  it "既にフォロー解除済みの場合でも200ステータスを返すこと" do
    user = create(:user)
    follower = create(:user)

    expect(follower.following?(user)).to be(false)

    login_as(follower, scope: :user)
    delete "/api/internal/follow", params: {user_id: user.id}

    expect(response.status).to eq(200)
    expect(follower.following?(user)).to be(false)
  end
end
