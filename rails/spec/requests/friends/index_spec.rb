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
end
