# frozen_string_literal: true

describe "GET /registrations/new", type: :request do
  context "when request is valid" do
    let(:email_confirmation) { create(:email_confirmation, event: "sign_up") }

    it "displays user registration page" do
      get "/registrations/new", params: {token: email_confirmation.token}

      expect(response.status).to eq(200)
      expect(response.body).to include("アカウント作成")
    end
  end
end
