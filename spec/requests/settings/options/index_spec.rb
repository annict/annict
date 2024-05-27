# typed: false
# frozen_string_literal: true

describe "GET /settings/options", type: :request do
  let!(:user) { create(:registered_user) }

  before do
    login_as(user, scope: :user)
  end

  it "ページが表示されること" do
    get "/settings/options"

    expect(response.status).to eq(200)
    expect(response.body).to include("未記録エピソードのネタバレを防ぐ")
  end
end
