# typed: false
# frozen_string_literal: true

RSpec.describe "GET /api/internal/people", type: :request do
  it "パラメータqが指定されていない場合、空の結果を返すこと" do
    get "/api/internal/people"
    expect(response).to have_http_status(:ok)
  end

  it "パラメータqが指定されている場合、マッチする人物を返すこと" do
    FactoryBot.create(:person, name: "山田太郎")
    FactoryBot.create(:person, name: "田中花子")
    FactoryBot.create(:person, name: "山田次郎")

    get "/api/internal/people", params: {q: "山田"}
    expect(response).to have_http_status(:ok)
  end

  it "パラメータqが空文字の場合、空の結果を返すこと" do
    FactoryBot.create(:person, name: "山田太郎")

    get "/api/internal/people", params: {q: ""}
    expect(response).to have_http_status(:ok)
  end

  it "大文字小文字を区別せずに検索できること" do
    FactoryBot.create(:person, name: "Yamada Taro")

    get "/api/internal/people", params: {q: "yamada"}
    expect(response).to have_http_status(:ok)

    get "/api/internal/people", params: {q: "YAMADA"}
    expect(response).to have_http_status(:ok)
  end

  it "部分一致で検索できること" do
    FactoryBot.create(:person, name: "声優山田太郎")

    get "/api/internal/people", params: {q: "山田"}
    expect(response).to have_http_status(:ok)
  end

  it "削除済みの人物は検索結果に含まれないこと" do
    person = FactoryBot.create(:person, name: "山田太郎")
    person.update!(deleted_at: Time.zone.now)

    get "/api/internal/people", params: {q: "山田"}
    expect(response).to have_http_status(:ok)
  end
end
