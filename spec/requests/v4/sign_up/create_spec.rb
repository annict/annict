# frozen_string_literal: true

describe "POST /sign_up", type: :request do
  context "when request is valid" do
    it "creates email confirmation" do
      expect(EmailConfirmation.count).to eq 0

      post "/sign_up", params: {
        sign_up_form: {
          email: "foo@example.com"
        }
      }

      expect(EmailConfirmation.count).to eq 1

      confirmation = EmailConfirmation.first
      expect(confirmation.email).to eq "foo@example.com"
      expect(confirmation.event).to eq "sign_up"
      expect(confirmation.token).to be_present

      expect(response).to redirect_to("/")
      expect(request.flash[:notice]).to include("アカウント作成のためのメールを送信しました")
    end
  end
end
