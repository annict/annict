# frozen_string_literal: true

describe "GET /settings/apps", type: :request do
  let!(:user) { create(:registered_user) }
  let!(:oauth_application) { create(:oauth_application) }
  let!(:oauth_access_token) { create(:oauth_access_token, application: oauth_application, owner: user) }

  before do
    login_as(user, scope: :user)
  end

  it "ページが表示されること" do
    get "/settings/apps"

    expect(response.status).to eq(200)
    expect(response.body).to include(oauth_application.name)
  end
end
