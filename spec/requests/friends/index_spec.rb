# typed: false
# frozen_string_literal: true

RSpec.describe "GET /friends", type: :request do
  it "ログインしているとき、アクセスできること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/friends"

    expect(response.status).to eq(200)
    expect(response.body).to include("SNSの友達")
  end

  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/friends"

    expect(response).to redirect_to(new_user_session_path)
  end

  it "Facebookとのセッションが切れているとき、設定ページにリダイレクトすること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    # social_friendsメソッドがall.pluckでエラーを発生させるように設定
    social_friends_mock = double("social_friends")
    allow(social_friends_mock).to receive(:all).and_return(social_friends_mock)
    allow(social_friends_mock).to receive(:pluck)
      .and_raise(Koala::Facebook::AuthenticationError.new(401, ""))
    allow(user).to receive(:social_friends).and_return(social_friends_mock)

    get "/friends"

    expect(response).to redirect_to(settings_provider_list_path)
    expect(flash[:alert]).to eq("Facebookとのセッションが切れました。再連携をしてください。")
  end
end
