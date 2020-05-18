# frozen_string_literal: true

describe "GET /", type: :request do
  before do
    host! "annict-jp.test:3000"
  end

  context "when user does not sign in" do
    let!(:work) { create(:work, :with_current_season) }

    it "displays welcome page" do
      get "/"

      expect(response.status).to eq(200)
      expect(response.body).to include("The platform for anime addicts.")
      expect(response.body).to include(work.title)
    end
  end
end
