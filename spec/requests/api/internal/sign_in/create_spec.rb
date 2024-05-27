# typed: false
# frozen_string_literal: true

describe "POST /api/internal/sign_in", type: :request do
  context "異常値を入力して送信したとき" do
    it "バリデーションエラーを返すこと" do
      expect(EmailConfirmation.count).to eq 0

      post "/api/internal/sign_in", params: {
        forms_sign_in_form: {
          email: "this-is-not-email"
        }
      }

      expect(EmailConfirmation.count).to eq 0

      expect(response.status).to eq(422)
      expect(response.body).to include("メールアドレスは不正な値です")
    end
  end

  context "正常値を入力して送信したとき" do
    let(:message_delivery) { instance_double(ActionMailer::MessageDelivery) }
    let(:user) { create :registered_user }

    before do
      allow(message_delivery).to receive(:deliver_later)
    end

    it "EmailConfirmation レコードが作成されて確認用メールが送信されること" do
      expect(EmailConfirmation.count).to eq 0

      expect(EmailConfirmationMailer).to receive(:sign_in_confirmation).and_return(message_delivery)
      expect(message_delivery).to receive(:deliver_later)

      post "/api/internal/sign_in", params: {
        forms_sign_in_form: {
          email: user.email
        }
      }

      expect(EmailConfirmation.count).to eq 1

      confirmation = EmailConfirmation.first
      expect(confirmation.email).to eq user.email
      expect(confirmation.event).to eq "sign_in"
      expect(confirmation.token).to be_present

      expect(response.status).to eq(201)
      expect(JSON.parse(response.body)).to eq({
        "flash" => {
          "type" => "notice",
          "message" => "ログインのためのメールを送信しました"
        }
      })
    end
  end
end
