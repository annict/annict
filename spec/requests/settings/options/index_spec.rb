# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/options", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/settings/options"

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログインしているとき、設定ページが正常に表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/options"

    expect(response.status).to eq(200)
    expect(response.body).to include("未記録エピソードのネタバレを防ぐ")
  end

  it "ログインしているとき、ユーザーの設定が表示されること" do
    user = create(:registered_user)
    user.setting.update!(hide_record_body: true, hide_supporter_badge: false)
    login_as(user, scope: :user)

    get "/settings/options"

    expect(response.status).to eq(200)
    expect(response.body).to include('checked="checked"')
  end
end
