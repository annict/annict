# frozen_string_literal: true

describe "GET /settings/account", type: :request do
  let!(:user) { create(:registered_user) }

  before do
    login_as(user, scope: :user)
  end

  it "ページが表示されること" do
    get "/settings/account"

    expect(response.status).to eq(200)
    expect(response.body).to include("基本情報")
  end
end
