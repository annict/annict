# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/search", type: :request do
  it "ログインしていないとき、リソースが保存されている場合、検索結果を表示すること" do
    series = create(:series, name: "しりーず検索")
    work = create(:work, title: "さくひん検索")
    person = create(:person, name: "じんぶつ検索")
    organization = create(:organization, name: "だんたい検索")
    character = create(:character, name: "きゃらくたー検索")

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
    expect(response.body).to include(work.title)
    expect(response.body).to include(person.name)
    expect(response.body).to include(organization.name)
    expect(response.body).to include(character.name)
  end

  it "ログインしていないとき、リソースが保存されていない場合、登録されていませんと表示すること" do
    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include("登録されていません")
  end

  it "ログインしているとき、リソースが保存されている場合、検索結果を表示すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    series = create(:series, name: "しりーず検索")
    work = create(:work, title: "さくひん検索")
    person = create(:person, name: "じんぶつ検索")
    organization = create(:organization, name: "だんたい検索")
    character = create(:character, name: "きゃらくたー検索")

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
    expect(response.body).to include(work.title)
    expect(response.body).to include(person.name)
    expect(response.body).to include(organization.name)
    expect(response.body).to include(character.name)
  end

  it "ログインしているとき、リソースが保存されていない場合、登録されていませんと表示すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include("登録されていません")
  end

  it "空の検索クエリの場合、正常にレスポンスを返すこと" do
    get "/db/search", params: {q: ""}

    expect(response.status).to eq(200)
  end

  it "検索クエリがnilの場合、正常にレスポンスを返すこと" do
    get "/db/search"

    expect(response.status).to eq(200)
  end

  it "ひらがな・カタカナの変換を考慮して検索結果を表示すること" do
    create(:series, name: "テストシリーズ")

    get "/db/search", params: {q: "てすと"}

    expect(response.status).to eq(200)
    expect(response.body).to include("登録されていません")
  end
end
