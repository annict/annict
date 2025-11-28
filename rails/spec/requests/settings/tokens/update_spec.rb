# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /settings/tokens/:token_id", type: :request do
  it "ログインしているとき、有効なパラメータでトークンを更新できること" do
    user = create(:registered_user)
    token = create(:oauth_access_token, resource_owner_id: user.id, application_id: nil, description: "Old Description", scopes: "read_anime")
    login_as(user, scope: :user)

    patch "/settings/tokens/#{token.id}", params: {
      oauth_access_token: {
        description: "Updated Description",
        scopes: "read_anime write_anime"
      }
    }

    expect(response.status).to eq(302)
    expect(response).to redirect_to(settings_app_list_path)
    expect(flash[:notice]).to be_present

    token.reload
    expect(token.description).to eq("Updated Description")
    expect(token.scopes.to_s).to eq("read_anime write_anime")
  end

  it "ログインしていないとき、ログインページにリダイレクトすること" do
    user = create(:registered_user)
    token = create(:oauth_access_token, resource_owner_id: user.id, application_id: nil, description: "Old Description", scopes: "read_anime")

    patch "/settings/tokens/#{token.id}", params: {
      oauth_access_token: {
        description: "Updated Description",
        scopes: "read_anime write_anime"
      }
    }

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)

    token.reload
    expect(token.description).to eq("Old Description")
    expect(token.scopes.to_s).to eq("read_anime")
  end

  it "他のユーザーのトークンを更新しようとした場合、NotFoundエラーになること" do
    user1 = create(:registered_user)
    user2 = create(:registered_user)
    token = create(:oauth_access_token, resource_owner_id: user1.id, application_id: nil, description: "Old Description", scopes: "read_anime")
    login_as(user2, scope: :user)

    expect do
      patch "/settings/tokens/#{token.id}", params: {
        oauth_access_token: {
          description: "Updated Description",
          scopes: "read_anime write_anime"
        }
      }
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないトークンIDを指定した場合、NotFoundエラーになること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect do
      patch "/settings/tokens/nonexistent", params: {
        oauth_access_token: {
          description: "Updated Description",
          scopes: "read_anime write_anime"
        }
      }
    end.to raise_error(ActiveRecord::RecordNotFound)
  end
end
