# typed: false
# frozen_string_literal: true

RSpec.describe "GET /api/internal/following", type: :request do
  it "未ログイン時は空の配列を返すこと" do
    get "/api/internal/following"

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq([])
  end

  it "ログイン時でフォロー中のユーザーがいない場合は空の配列を返すこと" do
    user = create(:user, :with_email_notification)

    login_as(user, scope: :user)
    get "/api/internal/following"

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq([])
  end

  it "ログイン時でフォロー中のユーザーがいる場合はユーザーIDの配列を返すこと" do
    user = create(:user, :with_email_notification)
    followed_user1 = create(:user, :with_email_notification)
    followed_user2 = create(:user, :with_email_notification)

    user.follow(followed_user1)
    user.follow(followed_user2)

    login_as(user, scope: :user)
    get "/api/internal/following"

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body).to match_array([followed_user1.id, followed_user2.id])
  end

  it "削除されたユーザーは結果に含まれないこと" do
    user = create(:user, :with_email_notification)
    followed_user1 = create(:user, :with_email_notification)
    followed_user2 = create(:user, :with_email_notification)

    user.follow(followed_user1)
    user.follow(followed_user2)
    followed_user2.update!(deleted_at: Time.current)

    login_as(user, scope: :user)
    get "/api/internal/following"

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body).to eq([followed_user1.id])
  end
end
