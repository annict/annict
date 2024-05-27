# typed: false
# frozen_string_literal: true

describe "GET /", type: :request do
  context "ログインしていないとき" do
    context "アニメが登録されていないとき" do
      it "Welcomeページが表示されること" do
        get "/"

        expect(response.status).to eq(200)
        expect(response.body).to include("A platform for anime addicts.")
        expect(response.body).to include("作品はありません")
      end
    end

    context "アニメが登録されているとき" do
      let!(:work) { create(:work, :with_current_season) }

      it "Welcomeページが表示されること" do
        get "/"

        expect(response.status).to eq(200)
        expect(response.body).to include("A platform for anime addicts.")
        expect(response.body).to include(work.title)
      end
    end
  end
end
