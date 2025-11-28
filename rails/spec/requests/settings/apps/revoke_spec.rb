# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /settings/apps/:app_id/revoke", type: :request do
  it "ログインしているユーザーがアプリケーションの接続を解除できること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application)
    oauth_access_token = FactoryBot.create(:oauth_access_token, application: oauth_application, owner: user)

    login_as(user, scope: :user)

    expect(oauth_access_token.revoked_at).to be_nil

    patch "/settings/apps/#{oauth_application.id}/revoke"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("接続解除しました")

    expect(oauth_access_token.reload.revoked_at).not_to be_nil
  end

  it "ログインしていないユーザーはアクセスできないこと" do
    oauth_application = FactoryBot.create(:oauth_application)

    patch "/settings/apps/#{oauth_application.id}/revoke"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "他のユーザーのアクセストークンは削除されないこと" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application)
    user_token = FactoryBot.create(:oauth_access_token, application: oauth_application, owner: user)
    other_user_token = FactoryBot.create(:oauth_access_token, application: oauth_application, owner: other_user)

    login_as(user, scope: :user)

    patch "/settings/apps/#{oauth_application.id}/revoke"

    expect(user_token.reload.revoked_at).not_to be_nil
    expect(other_user_token.reload.revoked_at).to be_nil
  end

  it "複数のアクセストークンがある場合、すべて削除されること" do
    user = FactoryBot.create(:registered_user)
    oauth_application = FactoryBot.create(:oauth_application)
    token1 = FactoryBot.create(:oauth_access_token, application: oauth_application, owner: user)
    token2 = FactoryBot.create(:oauth_access_token, application: oauth_application, owner: user)

    login_as(user, scope: :user)

    patch "/settings/apps/#{oauth_application.id}/revoke"

    expect(token1.reload.revoked_at).not_to be_nil
    expect(token2.reload.revoked_at).not_to be_nil
  end
end
