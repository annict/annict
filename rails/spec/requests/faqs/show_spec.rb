# typed: false
# frozen_string_literal: true

RSpec.describe "GET /faq", type: :request do
  it "GitHubに置いてあるドキュメントにリダイレクトすること" do
    get "/faq"

    expect(response).to redirect_to("https://github.com/annict/annict/blob/main/docs/faq.md")
  end
end
