# frozen_string_literal: true

describe "GET /", type: :request do
  before do
    host! "annict-jp.test:3000"
  end

  context "ログインしていないとき" do
    let!(:work) { create(:work, :with_current_season) }

    it "Welcomeページが表示されること" do
      get "/"

      expect(response.status).to eq(200)
      expect(response.body).to include("The platform for anime addicts.")
      expect(response.body).to include(work.title)
    end
  end
end
