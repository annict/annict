# typed: false
# frozen_string_literal: true

RSpec.describe "GET /userland", type: :request do
  it "カテゴリが存在しないとき、ステータス200でページが表示されること" do
    get "/userland"

    expect(response.status).to eq(200)
    expect(response.body).to include("Userland")
    expect(response.body).to include("Annict Userland")
  end

  it "カテゴリが存在するとき、ステータス200でカテゴリが表示されること" do
    UserlandCategory.create!(
      name: "テストカテゴリ",
      name_en: "Test Category",
      sort_number: 1
    )

    get "/userland"

    expect(response.status).to eq(200)
    expect(response.body).to include("テストカテゴリ")
    expect(response.body).to include("Userland")
  end

  it "複数のカテゴリが存在するとき、sort_number順で表示されること" do
    UserlandCategory.create!(
      name: "カテゴリ1",
      name_en: "Category 1",
      sort_number: 2
    )
    UserlandCategory.create!(
      name: "カテゴリ2",
      name_en: "Category 2",
      sort_number: 1
    )

    get "/userland"

    expect(response.status).to eq(200)
    expect(response.body).to include("カテゴリ1")
    expect(response.body).to include("カテゴリ2")

    # sort_number順で表示されているかチェック
    body = response.body
    category2_position = body.index("カテゴリ2")
    category1_position = body.index("カテゴリ1")
    expect(category2_position).to be < category1_position
  end

  it "レスポンスヘッダーに適切なContent-Typeが設定されていること" do
    get "/userland"

    expect(response.status).to eq(200)
    expect(response.headers["Content-Type"]).to include("text/html")
  end
end
