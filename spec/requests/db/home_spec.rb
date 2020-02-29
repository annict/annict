# frozen_string_literal: true

describe "Db::Home", type: :request do
  describe "GET /db" do
    it "has welcome message" do
      get "/db"

      expect(response.body).to include "Annict DBにようこそ！"
    end
  end
end
