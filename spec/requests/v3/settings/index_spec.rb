# frozen_string_literal: true

describe "GET /settings", type: :request do
  let!(:user) { create(:registered_user) }

  before do
    login_as(user, scope: :user)
  end

  it "ページが表示されること" do
    get "/settings", headers: { "User-Agent" => "Android" }

    expect(response.status).to eq(200)
    expect(response.body).to include("サービス連携")
  end
end
