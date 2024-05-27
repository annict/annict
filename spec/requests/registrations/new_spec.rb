# frozen_string_literal: true

describe "GET /registrations/new", type: :request do
  context "メールアドレス確認用トークンの有効期限が切れているとき" do
    let(:email_confirmation) { create(:email_confirmation, event: "sign_up", expires_at: 3.hours.ago) }

    it "有効期限切れのメッセージが表示されること" do
      get "/registrations/new", params: {token: email_confirmation.token}

      expect(response.status).to eq(200)
      expect(response.body).to include("アカウント作成用リンクの有効期限が切れました。")
    end
  end

  context "メールアドレス確認用トークンが正しいとき" do
    let(:email_confirmation) { create(:email_confirmation, event: "sign_up") }

    it "ユーザ登録フォームが表示されること" do
      get "/registrations/new", params: {token: email_confirmation.token}

      expect(response.status).to eq(200)
      expect(response.body).to include("アカウント作成")
    end
  end
end
