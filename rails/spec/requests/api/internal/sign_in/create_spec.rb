# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/sign_in", type: :request do
  it "メールアドレスが空のとき、バリデーションエラーを返すこと" do
    expect(EmailConfirmation.count).to eq 0

    post "/api/internal/sign_in", params: {
      forms_sign_in_form: {
        email: ""
      }
    }

    expect(EmailConfirmation.count).to eq 0

    expect(response.status).to eq(422)
    expect(response.body).to include("メールアドレスを入力してください")
  end

  it "メールアドレスが不正な形式のとき、バリデーションエラーを返すこと" do
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

  it "recaptchaトークンが検証に失敗したとき、エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    recaptcha = instance_double(Recaptcha)
    allow(Recaptcha).to receive(:new).with(action: "sign_in").and_return(recaptcha)
    allow(recaptcha).to receive(:verify?).and_return(false)

    expect(EmailConfirmation.count).to eq 0

    post "/api/internal/sign_in", params: {
      forms_sign_in_form: {
        email: user.email
      },
      recaptcha_token: "invalid_token"
    }

    expect(EmailConfirmation.count).to eq 0

    expect(response.status).to eq(422)
    expect(response.body).to include("reCAPTCHAの認証に失敗しました。時間を置いて再度お試しください")
  end

  it "正常なデータを送信したとき、EmailConfirmationレコードが作成されて確認用メールが送信されること" do
    user = FactoryBot.create(:registered_user)
    message_delivery = instance_double(ActionMailer::MessageDelivery)
    recaptcha = instance_double(Recaptcha)
    allow(Recaptcha).to receive(:new).with(action: "sign_in").and_return(recaptcha)
    allow(recaptcha).to receive(:verify?).and_return(true)
    allow(message_delivery).to receive(:deliver_later)

    expect(EmailConfirmation.count).to eq 0

    expect(EmailConfirmationMailer).to receive(:sign_in_confirmation).and_return(message_delivery)
    expect(message_delivery).to receive(:deliver_later)

    post "/api/internal/sign_in", params: {
      forms_sign_in_form: {
        email: user.email
      },
      recaptcha_token: "valid_token"
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

  it "メールアドレスの前後の空白を除去すること" do
    user = FactoryBot.create(:registered_user)
    message_delivery = instance_double(ActionMailer::MessageDelivery)
    recaptcha = instance_double(Recaptcha)
    allow(Recaptcha).to receive(:new).with(action: "sign_in").and_return(recaptcha)
    allow(recaptcha).to receive(:verify?).and_return(true)
    allow(message_delivery).to receive(:deliver_later)

    expect(EmailConfirmationMailer).to receive(:sign_in_confirmation).and_return(message_delivery)
    expect(message_delivery).to receive(:deliver_later)

    post "/api/internal/sign_in", params: {
      forms_sign_in_form: {
        email: "  #{user.email}  "
      },
      recaptcha_token: "valid_token"
    }

    expect(response.status).to eq(201)

    confirmation = EmailConfirmation.first
    expect(confirmation.email).to eq user.email
  end
end
