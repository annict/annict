# frozen_string_literal: true

describe "GET /faq", type: :request do
  it "GitHubに置いてあるドキュメントにリダイレクトすること" do
    get "/faq"

    expect(response).to redirect_to("https://github.com/kiraka/annict/blob/main/docs/faq.md")
  end
end
