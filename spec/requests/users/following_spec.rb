# typed: false
# frozen_string_literal: true

describe "GET /@:username/following", type: :request do
  let!(:user) { create(:registered_user) }

  context "フォロイーがいないとき" do
    it "アクセスできること" do
      get "/@#{user.username}/following"

      expect(response.status).to eq(200)
      expect(response.body).to include("フォローしていません")
    end
  end
end
