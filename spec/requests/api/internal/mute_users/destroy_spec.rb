# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /api/internal/mute_user", type: :request do
  it "認証していない場合、ログインページにリダイレクトされること" do
    user = create(:user)
    muted_user = create(:user)
    MuteUser.create!(user:, muted_user:)

    delete "/api/internal/mute_user", params: {user_id: muted_user.id}

    expect(response.status).to eq(302)
    expect(response).to redirect_to("/sign_in")
  end

  it "認証している場合、ミュートを解除できること" do
    user = create(:user, :with_email_notification)
    muted_user = create(:user)
    mute_user = MuteUser.create!(user:, muted_user:)

    login_as(user, scope: :user)

    expect(user.mute_users.count).to eq(1)

    delete "/api/internal/mute_user", params: {user_id: muted_user.id}

    expect(response.status).to eq(200)
    json = JSON.parse(response.body)
    expect(json["flash"]["type"]).to eq("notice")
    expect(json["flash"]["message"]).to include("ミュートを解除しました")

    expect(user.mute_users.count).to eq(0)
    expect(MuteUser.find_by(id: mute_user.id)).to be_nil
  end

  it "ミュートしていないユーザーのIDを指定した場合、エラーになること" do
    user = create(:user, :with_email_notification)
    not_muted_user = create(:user)

    login_as(user, scope: :user)

    expect {
      delete "/api/internal/mute_user", params: {user_id: not_muted_user.id}
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないユーザーIDを指定した場合、エラーになること" do
    user = create(:user, :with_email_notification)

    login_as(user, scope: :user)

    expect {
      delete "/api/internal/mute_user", params: {user_id: 999999}
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "user_idパラメータが指定されていない場合、エラーになること" do
    user = create(:user, :with_email_notification)

    login_as(user, scope: :user)

    expect {
      delete "/api/internal/mute_user"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "同じユーザーを複数回ミュート解除しようとした場合、2回目はエラーになること" do
    user = create(:user, :with_email_notification)
    muted_user = create(:user)
    MuteUser.create!(user:, muted_user:)

    login_as(user, scope: :user)

    # 1回目のミュート解除
    delete "/api/internal/mute_user", params: {user_id: muted_user.id}
    expect(response.status).to eq(200)

    # 2回目のミュート解除
    expect {
      delete "/api/internal/mute_user", params: {user_id: muted_user.id}
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
