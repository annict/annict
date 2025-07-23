# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/providers", type: :request do
  it "ログインしているときページが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/providers"

    expect(response.status).to eq(200)
    expect(response.body).to include("Googleカレンダー")
  end

  it "ログインしていないときログインページにリダイレクトされること" do
    get "/settings/providers"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end
end
