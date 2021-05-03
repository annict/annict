# frozen_string_literal: true

describe "GET /@:username/followers", type: :request do
  let!(:user) { create(:registered_user) }

  before do
    host! ENV.fetch("ANNICT_JP_HOST")
  end

  context "フォロワーがいないとき" do
    it "アクセスできること" do
      get "/@#{user.username}/followers"

      expect(response.status).to eq(200)
      expect(response.body).to include("フォローされていません")
    end
  end
end
