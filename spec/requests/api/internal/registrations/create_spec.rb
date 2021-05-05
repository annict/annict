# frozen_string_literal: true

describe "POST /api/internal/registrations", type: :request do
  context "異常値を入力して送信したとき" do
    let(:email_confirmation) { create(:email_confirmation, user: nil, event: "sign_up", back: "/welcome") }

    it "バリデーションエラーを返すこと" do
      expect(User.count).to eq 0

      post "/api/internal/registrations", params: {
        registration_form: {
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
  end

  context "正常値を入力して送信したとき" do
    let(:email_confirmation) { create(:email_confirmation, user: nil, event: "sign_up", back: "/welcome") }

    it "ユーザを作成してリダイレクトすること" do
      expect(User.count).to eq 0

      post "/api/internal/registrations", params: {
        registration_form: {
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
  end
end
