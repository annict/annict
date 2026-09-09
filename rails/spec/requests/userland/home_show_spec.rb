# typed: false
# frozen_string_literal: true

RSpec.describe "GET /userland", type: :request do
  let(:developer_help_selector) { 'a[href="https://wikino.app/s/annict/topics/5"]' }

  it "カテゴリが存在しないとき、ステータス200でページが表示されること" do
    get "/userland"

    expect(response.status).to eq(200)
    expect(response.body).to include("Userland")
    expect(response.body).to include("Annict Userland")
  end

  it "開発者向けヘルプへのリンクを正しい配置と文言で表示すること" do
    get "/userland"

    document = Nokogiri::HTML(response.body)
    sidebar = document.at_css(".c-main-sidebar")
    sidebar_misc_list = sidebar.at_xpath(".//div[normalize-space(.)='Misc']/following-sibling::ul[1]")
    sidebar_services_list = sidebar.at_xpath(".//div[normalize-space(.)='サービス']/following-sibling::ul[1]")
    footer = document.at_css(".c-footer")
    footer_contents_list = footer.at_xpath(".//h6[normalize-space(.)='コンテンツ']/following-sibling::ul[1]")
    footer_services_list = footer.at_xpath(".//h6[normalize-space(.)='サービス']/following-sibling::ul[1]")
    content = document.at_css(".l-default__content")

    expect(sidebar_misc_list.at_css(developer_help_selector).text.strip).to eq("開発者向けヘルプ")
    expect(sidebar_services_list.at_css(developer_help_selector)).to be_nil
    expect(footer_contents_list.at_css(developer_help_selector).text.strip).to eq("開発者向けヘルプ")
    expect(footer_services_list.at_css(developer_help_selector)).to be_nil
    expect(content.at_css(developer_help_selector).text.strip).to eq("Annict API")
    expect(document.css('a[href^="https://developers.annict.com"]')).to be_empty
  end

  it "英語ドメインでは開発者向けヘルプの英語ラベルを表示すること" do
    host! ENV.fetch("ANNICT_EN_DOMAIN")

    get "/userland"

    document = Nokogiri::HTML(response.body)

    expect(document.at_css(".c-main-sidebar").at_css(developer_help_selector).text.strip).to eq("Developer Help")
    expect(document.at_css(".c-footer").at_css(developer_help_selector).text.strip).to eq("Developer Help")
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

    # Check that categories are displayed by sort_number.
    #
    # [Ja] sort_number 順でカテゴリが表示されていることを確認する。
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
