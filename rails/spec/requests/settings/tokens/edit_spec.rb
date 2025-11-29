# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/tokens/:token_id/edit", type: :request do
  it "ログインしているとき、自分のトークンの編集ページが表示されること" do
    user = create(:registered_user)
    token = create(:oauth_access_token, application_id: nil, resource_owner_id: user.id, description: "Test Token")
    login_as(user, scope: :user)

    get "/settings/tokens/#{token.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(token.description)
  end

  it "ログインしているとき、存在しないトークンの場合404エラーが発生すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect do
      get "/settings/tokens/non_existent_id/edit"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ログインしているとき、他人のトークンの場合404エラーが発生すること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    token = create(:oauth_access_token, application_id: nil, resource_owner_id: other_user.id, description: "Other Token")
    login_as(user, scope: :user)

    expect do
      get "/settings/tokens/#{token.id}/edit"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ログインしていないとき、ログインページにリダイレクトすること" do
    user = create(:registered_user)
    token = create(:oauth_access_token, application_id: nil, resource_owner_id: user.id, description: "Test Token")

    get "/settings/tokens/#{token.id}/edit"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end
end
