# frozen_string_literal: true

describe "POST /registrations", type: :request do
  context "when request is valid" do
    let(:email_confirmation) { create(:email_confirmation, user: nil, event: "sign_up", back: "/welcome") }

    it "displays user registration page" do
      expect(User.count).to eq 0

      post "/registrations", params: {
        registration_form: {
          email: email_confirmation.email,
          token: email_confirmation.token,
          username: "example",
          terms_and_privacy_policy_agreement: 1
        }
      }

      expect(response).to redirect_to("/welcome")
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
