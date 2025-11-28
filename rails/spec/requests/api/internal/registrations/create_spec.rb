# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/registrations", type: :request do
  it "異常値を入力して送信したとき、バリデーションエラーを返すこと" do
    email_confirmation = create(:email_confirmation, user: nil, event: "sign_up", back: "/welcome")

    expect(User.count).to eq 0

    post "/api/internal/registrations", params: {
      forms_registration_form: {
        email: email_confirmation.email,
        token: email_confirmation.token,
        username: "this-is-not-username"
      }
    }

    # バリデーションエラーになるのでユーザは作成されないはず
    expect(User.count).to eq 0

    expect(response.status).to eq(422)
    expect(response.body).to include("利用規約とプライバシーポリシーを入力してください")
    expect(response.body).to include("ユーザ名は不正な値です")
  end

  it "正常値を入力して送信したとき、ユーザを作成してリダイレクトすること" do
    email_confirmation = create(:email_confirmation, user: nil, event: "sign_up", back: "/welcome")

    expect(User.count).to eq 0

    post "/api/internal/registrations", params: {
      forms_registration_form: {
        email: email_confirmation.email,
        token: email_confirmation.token,
        username: "example",
        terms_and_privacy_policy_agreement: 1
      }
    }

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq("redirect_path" => "/welcome")
    expect(request.flash[:notice]).to include("アカウント作成が完了しました。Annictにようこそ！")

    expect(User.count).to eq 1

    user = User.first
    expect(user.username).to eq "example"
    expect(user.email).to eq email_confirmation.email
    expect(user.role).to eq 0
    expect(user.confirmed_at).to be_present
    expect(user.profile).to be_present
    expect(user.setting).to be_present
    expect(user.email_notification).to be_present
  end

  it "無効なトークンで送信したとき、ActiveRecord::RecordNotFoundが発生すること" do
    email_confirmation = create(:email_confirmation, user: nil, event: "sign_up", back: "/welcome")

    expect(User.count).to eq 0

    expect {
      post "/api/internal/registrations", params: {
        forms_registration_form: {
          email: email_confirmation.email,
          token: "invalid_token",
          username: "example",
          terms_and_privacy_policy_agreement: 1
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)

    expect(User.count).to eq 0
  end

  it "既に存在するメールアドレスで送信したとき、バリデーションエラーを返すこと" do
    create(:user, email: "existing@example.com")
    email_confirmation = create(:email_confirmation, user: nil, event: "sign_up", back: "/welcome", email: "existing@example.com")

    expect(User.count).to eq 1

    post "/api/internal/registrations", params: {
      forms_registration_form: {
        email: email_confirmation.email,
        token: email_confirmation.token,
        username: "newuser",
        terms_and_privacy_policy_agreement: 1
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("メールアドレスはすでに存在します")
    expect(User.count).to eq 1
  end

  it "既に存在するユーザ名で送信したとき、バリデーションエラーを返すこと" do
    create(:user, username: "existing")
    email_confirmation = create(:email_confirmation, user: nil, event: "sign_up", back: "/welcome")

    expect(User.count).to eq 1

    post "/api/internal/registrations", params: {
      forms_registration_form: {
        email: email_confirmation.email,
        token: email_confirmation.token,
        username: "existing",
        terms_and_privacy_policy_agreement: 1
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("ユーザ名はすでに存在します")
    expect(User.count).to eq 1
  end

  it "利用規約とプライバシーポリシーに同意しない場合、バリデーションエラーを返すこと" do
    email_confirmation = create(:email_confirmation, user: nil, event: "sign_up", back: "/welcome")

    expect(User.count).to eq 0

    post "/api/internal/registrations", params: {
      forms_registration_form: {
        email: email_confirmation.email,
        token: email_confirmation.token,
        username: "example",
        terms_and_privacy_policy_agreement: 0
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("利用規約とプライバシーポリシーを受諾してください")
    expect(User.count).to eq 0
  end

  it "長すぎるユーザ名で送信したとき、バリデーションエラーを返すこと" do
    email_confirmation = create(:email_confirmation, user: nil, event: "sign_up", back: "/welcome")

    expect(User.count).to eq 0

    post "/api/internal/registrations", params: {
      forms_registration_form: {
        email: email_confirmation.email,
        token: email_confirmation.token,
        username: "a" * 21,  # 21文字（制限は20文字）
        terms_and_privacy_policy_agreement: 1
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("ユーザ名は20文字以内で入力してください")
    expect(User.count).to eq 0
  end

  it "backパラメータが指定されている場合、そのパスにリダイレクトすること" do
    email_confirmation = create(:email_confirmation, user: nil, event: "sign_up", back: "/custom/path")

    post "/api/internal/registrations", params: {
      forms_registration_form: {
        email: email_confirmation.email,
        token: email_confirmation.token,
        username: "example",
        terms_and_privacy_policy_agreement: 1
      }
    }

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq("redirect_path" => "/custom/path")
  end

  it "backパラメータが空の場合、ルートパスにリダイレクトすること" do
    email_confirmation = create(:email_confirmation, user: nil, event: "sign_up", back: "")

    post "/api/internal/registrations", params: {
      forms_registration_form: {
        email: email_confirmation.email,
        token: email_confirmation.token,
        username: "example",
        terms_and_privacy_policy_agreement: 1
      }
    }

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq("redirect_path" => "/")
  end
end
