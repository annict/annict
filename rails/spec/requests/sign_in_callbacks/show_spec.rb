# typed: false
# frozen_string_literal: true

RSpec.describe "GET /sign_in/callback", type: :request do
  it "正しいトークンが提供されたとき、ユーザーがサインインしてリダイレクトされること" do
    user = create :registered_user
    email_confirmation = create(:email_confirmation, email: user.email, event: "sign_in", back: "/home")

    expect(EmailConfirmation.count).to eq 1

    get "/sign_in/callback", params: {token: email_confirmation.token}

    expect(EmailConfirmation.count).to eq 0
    expect(response).to redirect_to("/home")
    expect(request.flash[:notice]).to include("ログインしました")
  end

  it "トークンが提供されていないとき、ルートパスにリダイレクトされること" do
    get "/sign_in/callback"

    expect(response).to redirect_to(root_path)
  end

  it "無効なトークンが提供されたとき、エラーメッセージが表示されること" do
    get "/sign_in/callback", params: {token: "invalid_token"}

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("ログイン用リンクの有効期限が切れました")
  end

  it "期限切れのトークンが提供されたとき、エラーメッセージが表示されること" do
    user = create :registered_user
    email_confirmation = create(:email_confirmation,
      email: user.email,
      event: "sign_in",
      expires_at: 1.hour.ago)

    get "/sign_in/callback", params: {token: email_confirmation.token}

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("ログイン用リンクの有効期限が切れました")
  end

  it "backパラメータが設定されていないとき、ルートパスにリダイレクトされること" do
    user = create :registered_user
    email_confirmation = create(:email_confirmation, email: user.email, event: "sign_in", back: nil)

    get "/sign_in/callback", params: {token: email_confirmation.token}

    expect(response).to redirect_to(root_path)
    expect(request.flash[:notice]).to include("ログインしました")
  end

  it "未確認ユーザーの場合、確認処理が実行されること" do
    user = create :user, :with_profile, :with_provider, :with_setting, :with_email_notification
    email_confirmation = create(:email_confirmation, email: user.email, event: "sign_in", back: "/home")

    expect(user.confirmed?).to be false

    get "/sign_in/callback", params: {token: email_confirmation.token}

    user.reload
    expect(user.confirmed?).to be true
    expect(response).to redirect_to("/home")
  end
end
