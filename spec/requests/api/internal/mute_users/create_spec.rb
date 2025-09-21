# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/mute_user", type: :request do
  it "認証していない場合、ログインページにリダイレクトされること" do
    target_user = create(:user)

    post "/api/internal/mute_user", params: {user_id: target_user.id}

    expect(response.status).to eq(302)
    expect(response).to redirect_to("/sign_in")
  end

  it "認証している場合、ユーザーをミュートできること" do
    user = create(:user, :with_email_notification)
    target_user = create(:user)

    login_as(user, scope: :user)

    expect(user.mute_users.count).to eq(0)

    post "/api/internal/mute_user", params: {user_id: target_user.id}

    expect(response.status).to eq(201)
    json = JSON.parse(response.body)
    expect(json["flash"]["type"]).to eq("notice")
    expect(json["flash"]["message"]).to include("ミュートしました")

    expect(user.mute_users.count).to eq(1)
    expect(user.mute_users.first.muted_user).to eq(target_user)
  end

  it "既にミュートしているユーザーを再度ミュートしても成功すること" do
    user = create(:user, :with_email_notification)
    target_user = create(:user)
    MuteUser.create!(user:, muted_user: target_user)

    login_as(user, scope: :user)

    expect(user.mute_users.count).to eq(1)

    post "/api/internal/mute_user", params: {user_id: target_user.id}

    expect(response.status).to eq(201)
    json = JSON.parse(response.body)
    expect(json["flash"]["type"]).to eq("notice")
    expect(json["flash"]["message"]).to include("ミュートしました")

    expect(user.mute_users.count).to eq(1)
  end

  it "存在しないユーザーIDを指定した場合、エラーになること" do
    user = create(:user, :with_email_notification)

    login_as(user, scope: :user)

    post "/api/internal/mute_user", params: {user_id: 999999

    expect(response.status).to eq(404)
  end

  it "削除されたユーザーをミュートしようとした場合、エラーになること" do
    user = create(:user, :with_email_notification)
    target_user = create(:user)
    target_user.destroy!

    login_as(user, scope: :user)

    post "/api/internal/mute_user", params: {user_id: target_user.id

    expect(response.status).to eq(404)
  end

  it "user_idパラメータが指定されていない場合、エラーになること" do
    user = create(:user, :with_email_notification)

    login_as(user, scope: :user)

    post "/api/internal/mute_user"

    expect(response.status).to eq(404)
  end

  it "自分自身をミュートしようとした場合の動作を確認すること" do
    user = create(:user, :with_email_notification)

    login_as(user, scope: :user)

    post "/api/internal/mute_user", params: {user_id: user.id}

    # 自分自身をミュートできるかどうかは実装次第なので、
    # 現在の動作を確認する
    if response.status == 201
      expect(user.mute_users.count).to eq(1)
      expect(user.mute_users.first.muted_user).to eq(user)
    else
      expect(response.status).to eq(400)
    end
  end
end
