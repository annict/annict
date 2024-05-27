# typed: false
# frozen_string_literal: true

describe "GET /@:username/followers", type: :request do
  let!(:user) { create(:registered_user) }

  context "フォロワーがいないとき" do
    it "アクセスできること" do
      get "/@#{user.username}/followers"

      expect(response.status).to eq(200)
      expect(response.body).to include("フォローされていません")
    end
  end
end
