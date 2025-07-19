# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/sign_up", type: :request do
  it "異常値を入力して送信したとき、バリデーションエラーを返すこと" do
    expect(EmailConfirmation.count).to eq 0

    post "/api/internal/sign_up", params: {
      forms_sign_up_form: {
        email: "this-is-not-email"
      }
    }

    expect(EmailConfirmation.count).to eq 0

    expect(response.status).to eq(422)
    expect(response.body).to include("メールアドレスは不正な値です")
  end

  it "正常値を入力して送信したとき、EmailConfirmation レコードが作成されて確認用メールが送信されること" do
    message_delivery = instance_double(ActionMailer::MessageDelivery)

    expect(EmailConfirmation.count).to eq 0

    expect(EmailConfirmationMailer).to receive(:sign_up_confirmation).and_return(message_delivery)
    expect(message_delivery).to receive(:deliver_later)

    post "/api/internal/sign_up", params: {
      forms_sign_up_form: {
        email: "foo@example.com"
      }
    }

    expect(EmailConfirmation.count).to eq 1

    confirmation = EmailConfirmation.first
    expect(confirmation.email).to eq "foo@example.com"
    expect(confirmation.event).to eq "sign_up"
    expect(confirmation.token).to be_present

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq({
      "flash" => {
        "type" => "notice",
        "message" => "アカウント作成のためのメールを送信しました"
      }
    })
  end

  it "recaptchaトークンが無効なとき、エラーを返すこと" do
    recaptcha = instance_double(Recaptcha)
    allow(Recaptcha).to receive(:new).with(action: "sign_up").and_return(recaptcha)
    allow(recaptcha).to receive(:verify?).with(nil).and_return(false)

    expect(EmailConfirmation.count).to eq 0

    post "/api/internal/sign_up", params: {
      forms_sign_up_form: {
        email: "valid@example.com"
      }
    }

    expect(EmailConfirmation.count).to eq 0

    expect(response.status).to eq(422)
    expect(response.body).to include("reCAPTCHAの認証に失敗しました。時間を置いて再度お試しください")
  end

  it "メールアドレスが空のとき、エラーを返すこと" do
    expect(EmailConfirmation.count).to eq 0

    post "/api/internal/sign_up", params: {
      forms_sign_up_form: {
        email: ""
      }
    }

    expect(EmailConfirmation.count).to eq 0

    expect(response.status).to eq(422)
    expect(response.body).to include("メールアドレスを入力してください")
  end

  it "メールアドレスに前後の空白がある場合、トリムして処理されること" do
    message_delivery = instance_double(ActionMailer::MessageDelivery)
    recaptcha = instance_double(Recaptcha)
    allow(Recaptcha).to receive(:new).with(action: "sign_up").and_return(recaptcha)
    allow(recaptcha).to receive(:verify?).and_return(true)

    expect(EmailConfirmation.count).to eq 0

    expect(EmailConfirmationMailer).to receive(:sign_up_confirmation).and_return(message_delivery)
    expect(message_delivery).to receive(:deliver_later)

    post "/api/internal/sign_up", params: {
      forms_sign_up_form: {
        email: "  foo@example.com  "
      },
      recaptcha_token: "valid-token"
    }

    expect(EmailConfirmation.count).to eq 1

    confirmation = EmailConfirmation.first
    expect(confirmation.email).to eq "foo@example.com"
  end
end
