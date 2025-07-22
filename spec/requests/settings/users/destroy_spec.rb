# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /settings/user", type: :request do
  it "ログイン済みユーザーで、アクティブなOAuthアプリケーションがない場合、アカウントが削除されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect {
      delete "/settings/user"
    }.to change { user.reload.deleted_at }.from(nil)

    expect(response).to redirect_to(root_path)
    expect(flash[:notice]).to eq("退会しました。(´・ω;:..")
  end

  it "ログイン済みユーザーで、アクティブなOAuthアプリケーションがある場合、エラーメッセージが表示されてアカウントが削除されないこと" do
    user = create(:registered_user)
    create(:oauth_application, owner: user, deleted_at: nil)
    login_as(user, scope: :user)

    expect {
      delete "/settings/user"
    }.not_to change { user.reload.deleted_at }

    expect(response).to redirect_to(root_path)
    expect(flash[:alert]).to be_present
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    delete "/settings/user"

    expect(response).to redirect_to(new_user_session_path)
  end

  it "アカウント削除時にユーザーがログアウトされること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    delete "/settings/user"

    expect(controller.current_user).to be_nil
  end

  it "アカウント削除時にユーザー名とメールアドレスが匿名化されること" do
    user = create(:registered_user, username: "testuser", email: "test@example.com")
    login_as(user, scope: :user)

    delete "/settings/user"

    user.reload
    expect(user.username).not_to eq("testuser")
    expect(user.email).not_to eq("test@example.com")
    expect(user.email).to end_with("@example.com")
  end

  it "アカウント削除時にプロバイダー情報が削除されること" do
    user = create(:registered_user)
    create(:provider, user: user)
    login_as(user, scope: :user)

    expect {
      delete "/settings/user"
    }.to change { user.providers.count }.to(0)
  end
end
