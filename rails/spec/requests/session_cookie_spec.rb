# typed: false
# frozen_string_literal: true

# `login_as` はWardenのテストモードで次の1リクエストだけにユーザーを登録するため、
# 2回目以降のリクエストでログイン状態を維持できるかはセッションCookieがリクエストのホストへ送られるかで決まる。
RSpec.describe "セッションCookie", type: :request do
  it "ログイン後の2回目のリクエストでも、ログイン状態が維持されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/muted_users"
    expect(response.status).to eq(200)

    get "/settings/muted_users"
    expect(response.status).to eq(200)
  end

  it "セッションCookieにドメインを指定せず、リクエストのホストで保存されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/muted_users"

    session_cookie = Array(response.headers["Set-Cookie"]).join("\n").lines.find { |line| line.start_with?("_annict_session_v201904=") }
    expect(session_cookie).to be_present
    expect(session_cookie).not_to match(/domain=/i)
  end

  it "ログインしていない場合、ログインページにリダイレクトされること" do
    get "/settings/muted_users"

    expect(response).to redirect_to(new_user_session_path)
  end
end
