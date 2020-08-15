# frozen_string_literal: true

describe "POST /legacy/sign_in", type: :request do
  context "when request is valid" do
    let(:user) { create :registered_user }
    let(:password) { "password" }

    before do
      user.update(password: password)
    end

    it "signs in" do
      post "/legacy/sign_in", params: {
        user: {
          email_username: user.username,
          password: password
        }
      }

      expect(response).to redirect_to("/")
      expect(request.flash[:notice]).to include("ログインしました")
    end
  end
end
