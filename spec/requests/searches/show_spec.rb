# frozen_string_literal: true

describe "GET /search", type: :request do
  describe "アニメ検索" do
    let!(:anime) { create(:anime) }

    it "検索結果が表示できること" do
      get "/search", params: {q: anime.title}

      expect(response.status).to eq(200)
      expect(response.body).to include(anime.title)
    end
  end

  describe "キャラクター検索" do
    let!(:character) { create(:character) }

    it "検索結果が表示できること" do
      get "/search", params: {q: character.name}

      expect(response.status).to eq(200)
      expect(response.body).to include(character.name)
    end
  end

  describe "人物検索" do
    let!(:person) { create(:person) }

    it "検索結果が表示できること" do
      get "/search", params: {q: person.name}

      expect(response.status).to eq(200)
      expect(response.body).to include(person.name)
    end
  end

  describe "団体検索" do
    let!(:organization) { create(:organization) }

    it "検索結果が表示できること" do
      get "/search", params: {q: organization.name}

      expect(response.status).to eq(200)
      expect(response.body).to include(organization.name)
    end
  end
end
