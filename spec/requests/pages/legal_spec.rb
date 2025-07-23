# typed: false
# frozen_string_literal: true

RSpec.describe "GET /legal", type: :request do
  it "特商法ページが表示されること" do
    get "/legal"

    expect(response.status).to eq(200)
    expect(response.body).to include("サービス名")
    expect(response.body).to include("運営責任者")
    expect(response.body).to include("榛葉 光二")
    expect(response.body).to include("販売価格")
    expect(response.body).to include("月額290円または年額2,900円")
    expect(response.body).to include('<meta name="robots" content="noindex,nofollow,noarchive">')
  end
end
