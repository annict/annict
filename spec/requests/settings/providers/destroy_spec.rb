# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /settings/providers/:provider_id", type: :request do
  it "ログインしているとき、プロバイダーが削除されて元のページにリダイレクトされること" do
    user = create(:registered_user)
    provider = create(:provider, user:)
    login_as(user, scope: :user)

    delete "/settings/providers/#{provider.id}"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(settings_provider_list_path)
    expect(flash[:notice]).to eq(I18n.t("messages.providers.removed"))
    provider.reload
    expect(provider.deleted?).to eq(true)
  end

  it "ログインしているとき、他のユーザーのプロバイダーを削除しようとすると404エラーが返されること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    provider = create(:provider, user: other_user)
    login_as(user, scope: :user)

    delete "/settings/providers/#{provider.id}"

    expect(response).to have_http_status(:not_found)
  end

  it "ログインしているとき、存在しないプロバイダーを削除しようとすると404エラーが返されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)
    non_existent_id = "non_existent_id"

    delete "/settings/providers/#{non_existent_id}"

    expect(response).to have_http_status(:not_found)
  end

  it "ログインしていないときログインページにリダイレクトされること" do
    user = create(:registered_user)
    provider = create(:provider, user:)

    delete "/settings/providers/#{provider.id}"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログインしているとき、Refererヘッダーが設定されている場合はRefererにリダイレクトされること" do
    user = create(:registered_user)
    provider = create(:provider, user:)
    login_as(user, scope: :user)
    referer_url = "/some/custom/path"

    delete "/settings/providers/#{provider.id}", headers: {"HTTP_REFERER" => referer_url}

    expect(response.status).to eq(302)
    expect(response).to redirect_to(referer_url)
    expect(flash[:notice]).to eq(I18n.t("messages.providers.removed"))
  end
end
