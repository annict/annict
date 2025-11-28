# typed: false
# frozen_string_literal: true

RSpec.describe "POST /settings/tokens", type: :request do
  it "ログインしているとき、有効なパラメータでトークンを作成できること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    post "/settings/tokens", params: {
      oauth_access_token: {
        description: "Test Token",
        scopes: "read_anime"
      }
    }

    expect(response.status).to eq(302)
    expect(response).to redirect_to(settings_app_list_path)
    expect(flash[:notice]).to be_present
    expect(flash[:created_token]).to be_present
    expect(user.oauth_access_tokens.personal.count).to eq(1)

    token = user.oauth_access_tokens.personal.last
    expect(token.description).to eq("Test Token")
    expect(token.scopes.to_s).to eq("read_anime")
  end

  it "ログインしているとき、無効なパラメータの場合エラーページが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    post "/settings/tokens", params: {
      oauth_access_token: {
        description: "",
        scopes: ""
      }
    }

    expect(response.status).to eq(422)
    expect(user.oauth_access_tokens.personal.count).to eq(0)
  end

  it "ログインしていないとき、ログインページにリダイレクトすること" do
    post "/settings/tokens", params: {
      oauth_access_token: {
        description: "Test Token",
        scopes: "read_anime"
      }
    }

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end
end
