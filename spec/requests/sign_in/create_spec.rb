# frozen_string_literal: true

describe "POST /sign_in", type: :request do
  context "when request is valid" do
    let(:user) { create :registered_user }

    it "creates email confirmation" do
      expect(EmailConfirmation.count).to eq 0

      post "/sign_in", params: {
        sign_in_form: {
          email: user.email
        }
      }

      expect(EmailConfirmation.count).to eq 1

      confirmation = EmailConfirmation.first
      expect(confirmation.email).to eq user.email
      expect(confirmation.event).to eq "sign_in"
      expect(confirmation.token).to be_present

      expect(response).to redirect_to("/")
      expect(request.flash[:notice]).to include("ログインのためのメールを送信しました")
    end
  end
end
