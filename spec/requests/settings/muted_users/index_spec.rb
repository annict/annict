# frozen_string_literal: true

describe "GET /settings/muted_users", type: :request do
  let!(:user_1) { create(:registered_user) }
  let!(:user_2) { create(:registered_user) }

  before do
    user_1.mute(user_2)

    login_as(user_1, scope: :user)
  end

  it "ページが表示されること" do
    get "/settings/muted_users"

    expect(response.status).to eq(200)
    expect(response.body).to include(user_2.profile.name)
  end
end
