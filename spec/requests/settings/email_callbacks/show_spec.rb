# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/email/callback", type: :request do
  it "ログインしていない場合、ログインページにリダイレクトされること" do
    get "/settings/email/callback"

    expect(response).to redirect_to("/sign_in")
  end

  it "tokenパラメータがない場合、rootにリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/email/callback"

    expect(response).to redirect_to(root_path)
  end

  it "対応するconfirmationが存在しない場合、rootにリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/email/callback", params: {token: "invalid-token"}

    expect(response).to redirect_to(root_path)
  end

  it "confirmationが期限切れの場合、rootにリダイレクトされ、エラーメッセージが表示されること" do
    user = FactoryBot.create(:registered_user)
    confirmation = FactoryBot.create(
      :email_confirmation,
      user: user,
      event: "update_email",
      email: "new@example.com",
      expires_at: 1.hour.ago
    )
    login_as(user, scope: :user)

    get "/settings/email/callback", params: {token: confirmation.token}

    expect(response).to redirect_to(root_path)
    expect(flash[:alert]).to eq("メールアドレス確認リンクの有効期限が切れました。再度やり直してください")
  end

  it "有効なconfirmationがある場合、メールアドレスが更新され、confirmationが削除されること" do
    user = FactoryBot.create(:registered_user, email: "old@example.com")
    new_email = "new@example.com"
    confirmation = FactoryBot.create(
      :email_confirmation,
      user: user,
      event: "update_email",
      email: new_email,
      expires_at: 1.hour.from_now
    )
    login_as(user, scope: :user)

    expect {
      get "/settings/email/callback", params: {token: confirmation.token}
    }.to change { user.reload.email }.from("old@example.com").to(new_email)
      .and change(EmailConfirmation, :count).by(-1)

    expect(response).to redirect_to(root_path)
    expect(flash[:notice]).to eq("メールアドレスを更新しました")
  end

  it "他のユーザーのconfirmationトークンを使用した場合、rootにリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    other_user = FactoryBot.create(:registered_user)
    confirmation = FactoryBot.create(
      :email_confirmation,
      user: other_user,
      event: "update_email",
      email: "new@example.com",
      expires_at: 1.hour.from_now
    )
    login_as(user, scope: :user)

    get "/settings/email/callback", params: {token: confirmation.token}

    expect(response).to redirect_to(root_path)
    expect(user.reload.email).not_to eq("new@example.com")
  end
end
