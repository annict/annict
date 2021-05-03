# frozen_string_literal: true

describe "GET /sign_in/callback", type: :request do
  context "when request is valid" do
    let(:user) { create :registered_user }
    let!(:email_confirmation) { create(:email_confirmation, email: user.email, event: "sign_in", back: "/home") }

    it "signs in" do
      expect(EmailConfirmation.count).to eq 1

      get "/sign_in/callback", params: {token: email_confirmation.token}

      expect(EmailConfirmation.count).to eq 0

      expect(response).to redirect_to("/home")
      expect(request.flash[:notice]).to include("ログインしました")
    end
  end
end
