# frozen_string_literal: true

describe "GET /sign_in", type: :request do
  context "ログインしているとき" do
    let!(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    it "トップページにリダイレクトされること" do
      get "/sign_in"

      expect(response.status).to eq(302)
      expect(response).to redirect_to("/")
    end
  end

  context "ログインしていないとき" do
    it "ログインページが表示されること" do
      get "/sign_in"

      expect(response.status).to eq(200)
      expect(response.body).to include("おかえりなさい！")
    end
  end
end
