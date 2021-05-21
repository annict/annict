# frozen_string_literal: true

describe "GET /settings/apps/:app_id/revoke", type: :request do
  let!(:user) { create(:registered_user) }
  let!(:oauth_application) { create(:oauth_application) }
  let!(:oauth_access_token) { create(:oauth_access_token, application: oauth_application, owner: user) }

  before do
    login_as(user, scope: :user)
  end

  it "接続解除できること" do
    expect(oauth_access_token.revoked_at).to be_nil

    patch "/settings/apps/#{oauth_application.id}/revoke"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("接続解除しました")

    expect(oauth_access_token.reload.revoked_at).not_to be_nil
  end
end
