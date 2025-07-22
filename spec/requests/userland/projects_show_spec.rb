# typed: false
# frozen_string_literal: true

RSpec.describe "GET /userland/projects/:project_id", type: :request do
  it "存在するプロジェクトの場合、ステータス200でプロジェクトが表示されること" do
    category = FactoryBot.create(:userland_category)
    project = FactoryBot.create(:userland_project, userland_category: category)

    get "/userland/projects/#{project.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(project.name)
    expect(response.body).to include(project.summary)
  end

  it "存在しないプロジェクトの場合、404エラーが発生すること" do
    expect {
      get "/userland/projects/999999"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "プロジェクトの詳細情報が表示されること" do
    category = FactoryBot.create(:userland_category, name: "テストカテゴリ")
    project = FactoryBot.create(
      :userland_project,
      userland_category: category,
      name: "テストプロジェクト",
      summary: "テストプロジェクトの概要",
      description: "テストプロジェクトの詳細説明",
      url: "https://test-project.com"
    )

    get "/userland/projects/#{project.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include("テストプロジェクト")
    expect(response.body).to include("テストプロジェクトの概要")
    expect(response.body).to include("https://test-project.com")
  end

  it "レスポンスヘッダーに適切なContent-Typeが設定されていること" do
    category = FactoryBot.create(:userland_category)
    project = FactoryBot.create(:userland_project, userland_category: category)

    get "/userland/projects/#{project.id}"

    expect(response.status).to eq(200)
    expect(response.headers["Content-Type"]).to include("text/html")
  end

  it "利用不可のプロジェクトでもアクセス可能であること" do
    category = FactoryBot.create(:userland_category)
    project = FactoryBot.create(
      :userland_project,
      userland_category: category,
      available: false
    )

    get "/userland/projects/#{project.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(project.name)
  end
end
