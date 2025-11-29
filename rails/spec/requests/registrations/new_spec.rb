# typed: false
# frozen_string_literal: true

RSpec.describe "GET /registrations/new", type: :request do
  it "メールアドレス確認用トークンの有効期限が切れているとき、有効期限切れのメッセージが表示されること" do
    email_confirmation = create(:email_confirmation, event: "sign_up", expires_at: 3.hours.ago)

    get "/registrations/new", params: {token: email_confirmation.token}

    expect(response.status).to eq(200)
    expect(response.body).to include("アカウント作成用リンクの有効期限が切れました。")
  end

  it "メールアドレス確認用トークンが正しいとき、ユーザ登録フォームが表示されること" do
    email_confirmation = create(:email_confirmation, event: "sign_up")

    get "/registrations/new", params: {token: email_confirmation.token}

    expect(response.status).to eq(200)
    expect(response.body).to include("アカウント作成")
  end

  it "トークンが指定されていないとき、ルートページにリダイレクトされること" do
    get "/registrations/new"

    expect(response).to redirect_to(root_path)
  end

  it "存在しないトークンが指定されたとき、有効期限切れのメッセージが表示されること" do
    get "/registrations/new", params: {token: "invalid-token"}

    expect(response.status).to eq(200)
    expect(response.body).to include("アカウント作成用リンクの有効期限が切れました。")
  end
end
