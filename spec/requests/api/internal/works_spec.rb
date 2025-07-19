# typed: false
# frozen_string_literal: true

RSpec.describe "GET /api/internal/works", type: :request do
  it "クエリパラメータqがある場合、タイトルが部分一致する作品を返すこと" do
    work1 = FactoryBot.create(:work, title: "進撃の巨人")
    work2 = FactoryBot.create(:work, title: "鋼の錬金術師")
    work3 = FactoryBot.create(:work, title: "進撃のバハムート")
    FactoryBot.create(:work, title: "ワンピース")

    get internal_api_work_list_path(q: "進撃")

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(work1.title)
    expect(response.body).to include(work3.title)
    expect(response.body).not_to include(work2.title)
    expect(response.body).not_to include("ワンピース")
  end

  it "クエリパラメータqがない場合、空の結果を返すこと" do
    FactoryBot.create(:work, title: "進撃の巨人")

    get internal_api_work_list_path

    expect(response).to have_http_status(:ok)
    expect(response.body).not_to include("進撃の巨人")
  end

  it "削除された作品を除外すること" do
    work1 = FactoryBot.create(:work, title: "進撃の巨人")
    work2 = FactoryBot.create(:work, title: "進撃のバハムート", deleted_at: Time.current)

    get internal_api_work_list_path(q: "進撃")

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(work1.title)
    expect(response.body).not_to include(work2.title)
  end

  it "大文字小文字を区別せずに検索すること" do
    work1 = FactoryBot.create(:work, title: "Attack on Titan")
    work2 = FactoryBot.create(:work, title: "ATTACK ON TITAN")
    work3 = FactoryBot.create(:work, title: "attack on titan")

    get internal_api_work_list_path(q: "attack")

    expect(response).to have_http_status(:ok)
    expect(response.body).to include(work1.title)
    expect(response.body).to include(work2.title)
    expect(response.body).to include(work3.title)
  end
end
