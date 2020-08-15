# frozen_string_literal: true

describe "GET /sign_in", type: :request do
  context "when request is valid" do
    it "displays sign in page" do
      get "/sign_in"

      expect(response.status).to eq(200)
      expect(response.body).to include("おかえりなさい！")
    end
  end
end
