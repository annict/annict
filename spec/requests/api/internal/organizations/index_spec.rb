# typed: false
# frozen_string_literal: true

RSpec.describe "GET /api/internal/organizations", type: :request do
  it "パラメータqが指定されていない場合、空の結果を返すこと" do
    get "/api/internal/organizations"
    expect(response).to have_http_status(:ok)
  end

  it "パラメータqが指定されている場合、マッチする組織を返すこと" do
    FactoryBot.create(:organization, name: "テスト制作会社1")
    FactoryBot.create(:organization, name: "サンプル制作会社")
    FactoryBot.create(:organization, name: "テスト制作会社2")

    get "/api/internal/organizations", params: {q: "テスト"}
    expect(response).to have_http_status(:ok)
  end

  it "パラメータqが空文字の場合、空の結果を返すこと" do
    FactoryBot.create(:organization, name: "テスト制作会社")

    get "/api/internal/organizations", params: {q: ""}
    expect(response).to have_http_status(:ok)
  end

  it "大文字小文字を区別せずに検索できること" do
    FactoryBot.create(:organization, name: "TEST Studio")

    get "/api/internal/organizations", params: {q: "test"}
    expect(response).to have_http_status(:ok)

    get "/api/internal/organizations", params: {q: "TEST"}
    expect(response).to have_http_status(:ok)
  end

  it "部分一致で検索できること" do
    FactoryBot.create(:organization, name: "東映アニメーション株式会社")

    get "/api/internal/organizations", params: {q: "アニメーション"}
    expect(response).to have_http_status(:ok)
  end

  it "削除済みの組織は検索結果に含まれないこと" do
    organization = FactoryBot.create(:organization, name: "テスト制作会社")
    organization.update!(deleted_at: Time.zone.now)

    get "/api/internal/organizations", params: {q: "テスト"}
    expect(response).to have_http_status(:ok)
  end
end
