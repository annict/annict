# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /settings/tokens/:token_id", type: :request do
  it "ログインしているとき、トークンが削除されて設定アプリ一覧ページにリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    token = FactoryBot.create(:oauth_access_token, resource_owner_id: user.id, application_id: nil, description: "個人トークン")
    login_as(user, scope: :user)

    delete "/settings/tokens/#{token.id}"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(settings_app_list_path)
    expect(flash[:notice]).to eq(I18n.t("messages.settings.tokens.deleted"))
    expect(Doorkeeper::AccessToken.exists?(id: token.id)).to eq(false)
  end

  it "ログインしているとき、他のユーザーのトークンを削除しようとすると404エラーが返されること" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    token = FactoryBot.create(:oauth_access_token, resource_owner_id: other_user.id, application_id: nil, description: "他ユーザーのトークン")
    login_as(user, scope: :user)

    delete "/settings/tokens/#{token.id}"

    expect(response).to have_http_status(:not_found)
  end

  it "ログインしているとき、存在しないトークンを削除しようとすると404エラーが返されること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)
    non_existent_id = "non_existent_id"

    delete "/settings/tokens/#{non_existent_id}"

    expect(response).to have_http_status(:not_found)
  end

  it "ログインしていないときログインページにリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    token = FactoryBot.create(:oauth_access_token, resource_owner_id: user.id, application_id: nil, description: "個人トークン")

    delete "/settings/tokens/#{token.id}"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログインしているとき、revoked（無効化）されたトークンを削除しようとすると404エラーが返されること" do
    user = FactoryBot.create(:registered_user)
    token = FactoryBot.create(:oauth_access_token, resource_owner_id: user.id, application_id: nil, description: "無効化されたトークン", revoked_at: Time.current)
    login_as(user, scope: :user)

    delete "/settings/tokens/#{token.id}"

    expect(response).to have_http_status(:not_found)
  end

  it "ログインしているとき、application_idが設定されているトークンを削除しようとすると404エラーが返されること" do
    user = FactoryBot.create(:registered_user)
    application = FactoryBot.create(:oauth_application)
    token = FactoryBot.create(:oauth_access_token, resource_owner_id: user.id, application:, description: "アプリケーショントークン")
    login_as(user, scope: :user)

    delete "/settings/tokens/#{token.id}"

    expect(response).to have_http_status(:not_found)
  end
end
