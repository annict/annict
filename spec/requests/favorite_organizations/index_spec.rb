# frozen_string_literal: true

describe "GET /@:username/favorite_organizations", type: :request do
  let!(:user) { create(:registered_user) }

  context "お気に入りがいないとき" do
    it "アクセスできること" do
      get "/@#{user.username}/favorite_organizations"

      expect(response.status).to eq(200)
      expect(response.body).to include("団体はありません")
    end
  end
end
