# typed: false
# frozen_string_literal: true

RSpec.describe "GET /api/internal/characters", type: :request do
  it "パラメータqが指定されていない場合、空の結果を返すこと" do
    get "/api/internal/characters"
    expect(response).to have_http_status(:ok)
  end

  it "パラメータqが指定されている場合、マッチするキャラクターを返すこと" do
    FactoryBot.create(:character, name: "テストキャラクター1")
    FactoryBot.create(:character, name: "サンプルキャラクター")
    FactoryBot.create(:character, name: "テストキャラクター2")

    get "/api/internal/characters", params: {q: "テスト"}
    expect(response).to have_http_status(:ok)
  end

  it "パラメータqが空文字の場合、空の結果を返すこと" do
    FactoryBot.create(:character, name: "テストキャラクター")

    get "/api/internal/characters", params: {q: ""}
    expect(response).to have_http_status(:ok)
  end

  it "大文字小文字を区別せずに検索できること" do
    FactoryBot.create(:character, name: "テストキャラクター")

    get "/api/internal/characters", params: {q: "テスト"}
    expect(response).to have_http_status(:ok)

    get "/api/internal/characters", params: {q: "てすと"}
    expect(response).to have_http_status(:ok)
  end

  it "部分一致で検索できること" do
    FactoryBot.create(:character, name: "美少女戦士セーラームーン")

    get "/api/internal/characters", params: {q: "戦士"}
    expect(response).to have_http_status(:ok)
  end

  it "削除済みのキャラクターは検索結果に含まれないこと" do
    character = FactoryBot.create(:character, name: "テストキャラクター")
    character.update!(deleted_at: Time.zone.now)

    get "/api/internal/characters", params: {q: "テスト"}
    expect(response).to have_http_status(:ok)
  end
end
