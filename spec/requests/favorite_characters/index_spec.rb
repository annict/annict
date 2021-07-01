# frozen_string_literal: true

describe "GET /@:username/favorite_characters", type: :request do
  let!(:user) { create(:registered_user) }

  context "お気に入りがいないとき" do
    it "アクセスできること" do
      get "/@#{user.username}/favorite_characters"

      expect(response.status).to eq(200)
      expect(response.body).to include("キャラクターはいません")
    end
  end
end
