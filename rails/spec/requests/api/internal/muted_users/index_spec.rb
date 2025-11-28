# typed: false
# frozen_string_literal: true

RSpec.describe "GET /api/internal/muted_users", type: :request do
  it "認証していない場合、空の配列が返されること" do
    get "/api/internal/muted_users"

    expect(response.status).to eq(200)
    json = JSON.parse(response.body)
    expect(json).to eq([])
  end

  it "認証している場合でミュートしているユーザーがいない場合、空の配列が返されること" do
    user = create(:user)

    login_as(user, scope: :user)

    get "/api/internal/muted_users"

    expect(response.status).to eq(200)
    json = JSON.parse(response.body)
    expect(json).to eq([])
  end

  it "認証している場合でミュートしているユーザーがいる場合、ミュートしているユーザーのIDの配列が返されること" do
    user = create(:user)
    muted_user1 = create(:user)
    muted_user2 = create(:user)
    muted_user3 = create(:user)

    # user が muted_user1 と muted_user2 をミュート
    MuteUser.create!(user:, muted_user: muted_user1)
    MuteUser.create!(user:, muted_user: muted_user2)

    login_as(user, scope: :user)

    get "/api/internal/muted_users"

    expect(response.status).to eq(200)
    json = JSON.parse(response.body)
    expect(json).to match_array([muted_user1.id, muted_user2.id])
    expect(json).not_to include(muted_user3.id)
  end

  it "他のユーザーがミュートしているユーザーは含まれないこと" do
    user1 = create(:user)
    user2 = create(:user)
    muted_user1 = create(:user)
    muted_user2 = create(:user)

    # user1 が muted_user1 をミュート
    MuteUser.create!(user: user1, muted_user: muted_user1)
    # user2 が muted_user2 をミュート
    MuteUser.create!(user: user2, muted_user: muted_user2)

    login_as(user1, scope: :user)

    get "/api/internal/muted_users"

    expect(response.status).to eq(200)
    json = JSON.parse(response.body)
    expect(json).to eq([muted_user1.id])
    expect(json).not_to include(muted_user2.id)
  end

  it "削除されたユーザーをミュートしている場合、そのミュート関係も削除されること" do
    user = create(:user)
    muted_user = create(:user)

    MuteUser.create!(user:, muted_user:)
    expect(user.mute_users.count).to eq(1)

    muted_user.destroy!

    login_as(user, scope: :user)

    get "/api/internal/muted_users"

    expect(response.status).to eq(200)
    json = JSON.parse(response.body)
    expect(json).to eq([])
    expect(user.mute_users.count).to eq(0)
  end
end
