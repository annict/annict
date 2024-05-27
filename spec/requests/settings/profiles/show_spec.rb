# typed: false
# frozen_string_literal: true

describe "GET /settings/profile", type: :request do
  let!(:user) { create(:registered_user) }

  before do
    login_as(user, scope: :user)
  end

  it "ページが表示されること" do
    get "/settings/profile"

    expect(response.status).to eq(200)
    expect(response.body).to include("自己紹介")
  end
end
