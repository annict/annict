# typed: false
# frozen_string_literal: true

RSpec.describe "GET|POST /users/auth/gumroad", type: :request do
  it "ログインしていない場合、Gumroadの認証を開始すること" do
    post "/users/auth/gumroad"

    expect(response).to have_http_status(:found)
  end

  it "ログインしている場合、Gumroadの認証を開始すること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    post "/users/auth/gumroad"

    expect(response).to have_http_status(:found)
  end
end
