# typed: false
# frozen_string_literal: true

RSpec.describe "GET /search", type: :request do
  it "アニメを検索できること" do
    work = FactoryBot.create(:work)

    get "/search", params: {q: work.title}

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "キャラクターを検索できること" do
    character = FactoryBot.create(:character)

    get "/search", params: {q: character.name}

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
  end

  it "人物を検索できること" do
    person = FactoryBot.create(:person)

    get "/search", params: {q: person.name}

    expect(response.status).to eq(200)
    expect(response.body).to include(person.name)
  end

  it "団体を検索できること" do
    organization = FactoryBot.create(:organization)

    get "/search", params: {q: organization.name}

    expect(response.status).to eq(200)
    expect(response.body).to include(organization.name)
  end

  it "検索キーワードが空の場合でも正常に表示されること" do
    get "/search", params: {q: ""}

    expect(response.status).to eq(200)
  end

  it "検索キーワードがない場合でも正常に表示されること" do
    get "/search"

    expect(response.status).to eq(200)
  end

  it "デフォルトでアニメの検索結果が表示されること" do
    work = FactoryBot.create(:work, title: "デフォルト検索アニメ")
    character = FactoryBot.create(:character, name: "デフォルト検索キャラ")

    get "/search", params: {q: "デフォルト検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
    expect(response.body).not_to include(character.name)
  end

  it "リソースタイプを指定して検索できること" do
    work = FactoryBot.create(:work, title: "リソース検索アニメ")
    character = FactoryBot.create(:character, name: "リソース検索キャラ")

    # アニメを検索
    get "/search", params: {q: "リソース検索", resource: "work"}
    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)

    # キャラクターを検索
    get "/search", params: {q: "リソース検索", resource: "character"}
    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
  end
end
