# frozen_string_literal: true

describe "GET /@:username/favorite_characters", type: :request do
  let!(:user) { create(:registered_user) }

  before do
    host! ENV.fetch("ANNICT_JP_HOST")
  end

  context "お気に入りがいないとき" do
    it "アクセスできること" do
      get "/@#{user.username}/favorite_characters"

      expect(response.status).to eq(200)
      expect(response.body).to include("キャラクターはいません")
    end
  end
end
