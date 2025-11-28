# typed: false
# frozen_string_literal: true

RSpec.describe "GET /api/internal/series_list", type: :request do
  it "パラメータqが無い場合は空の配列を返すこと" do
    create(:series, name: "ガンダムシリーズ")
    create(:series, name: "プリキュアシリーズ")

    get "/api/internal/series_list"

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body["resources"]).to eq([])
  end

  it "パラメータqに一致するシリーズを返すこと" do
    series1 = create(:series, name: "ガンダムシリーズ")
    create(:series, name: "プリキュアシリーズ")
    series3 = create(:series, name: "ガンダムビルドファイターズシリーズ")

    get "/api/internal/series_list", params: {q: "ガンダム"}

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body["resources"].size).to eq(2)
    expect(response_body["resources"]).to contain_exactly(
      {"id" => series1.id, "text" => "ガンダムシリーズ"},
      {"id" => series3.id, "text" => "ガンダムビルドファイターズシリーズ"}
    )
  end

  it "大文字小文字を区別せずに検索すること" do
    series1 = create(:series, name: "GUNDAM")
    series2 = create(:series, name: "gundam")
    series3 = create(:series, name: "Gundam")

    get "/api/internal/series_list", params: {q: "gUnDaM"}

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body["resources"].size).to eq(3)
    expect(response_body["resources"]).to contain_exactly(
      {"id" => series1.id, "text" => "GUNDAM"},
      {"id" => series2.id, "text" => "gundam"},
      {"id" => series3.id, "text" => "Gundam"}
    )
  end

  it "部分一致で検索すること" do
    series1 = create(:series, name: "機動戦士ガンダム")
    series2 = create(:series, name: "ガンダムSEED")
    series3 = create(:series, name: "SDガンダムフォース")
    create(:series, name: "プリキュア")

    get "/api/internal/series_list", params: {q: "ガンダム"}

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body["resources"].size).to eq(3)
    expect(response_body["resources"]).to contain_exactly(
      {"id" => series1.id, "text" => "機動戦士ガンダム"},
      {"id" => series2.id, "text" => "ガンダムSEED"},
      {"id" => series3.id, "text" => "SDガンダムフォース"}
    )
  end

  it "削除済みのシリーズは除外すること" do
    series1 = create(:series, name: "ガンダムシリーズ")
    create(:series, :deleted, name: "削除済みガンダムシリーズ")

    get "/api/internal/series_list", params: {q: "ガンダム"}

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body["resources"].size).to eq(1)
    expect(response_body["resources"]).to eq([
      {"id" => series1.id, "text" => "ガンダムシリーズ"}
    ])
  end

  it "空文字列のパラメータqの場合は空の配列を返すこと" do
    create(:series, name: "ガンダムシリーズ")

    get "/api/internal/series_list", params: {q: ""}

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body["resources"]).to eq([])
  end

  it "ローカライズされた名前を返すこと" do
    series = create(:series, name: "ガンダムシリーズ", name_en: "Gundam Series")

    get "/api/internal/series_list", params: {q: "ガンダム"}

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body["resources"]).to eq([
      {"id" => series.id, "text" => "ガンダムシリーズ"}
    ])
  end

  it "特殊文字を含む検索クエリを適切に処理すること" do
    series = create(:series, name: "ガンダム%シリーズ")

    get "/api/internal/series_list", params: {q: "%"}

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body["resources"]).to eq([
      {"id" => series.id, "text" => "ガンダム%シリーズ"}
    ])
  end
end
