# typed: false
# frozen_string_literal: true

describe "GraphQL API Query" do
  describe "searchCharacters" do
    let!(:character1) { create(:character, favorite_users_count: 10) }
    let!(:character2) { create(:character, favorite_users_count: 30) }
    let!(:character3) { create(:character, favorite_users_count: 20) }

    context "when `annictIds` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchCharacters(annictIds: [#{character1.id}]) {
              edges {
                node {
                  name
                }
              }
            }
          }
        QUERY

        res = Beta::AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows character name" do
        expect(result.dig("data", "searchCharacters", "edges")).to match_array([
          {
            "node" => {
              "name" => character1.name
            }
          }
        ])
      end
    end

    context "when `names` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchCharacters(names: ["#{character3.name}"]) {
              edges {
                node {
                  name
                }
              }
            }
          }
        QUERY

        res = Beta::AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows character name" do
        expect(result.dig("data", "searchCharacters", "edges")).to match_array([
          {
            "node" => {
              "name" => character3.name
            }
          }
        ])
      end
    end

    context "when `orderBy` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchCharacters(orderBy: { field: FAVORITE_CHARACTERS_COUNT, direction: DESC }) {
              edges {
                node {
                  name
                  favoriteCharactersCount
                }
              }
            }
          }
        QUERY

        res = Beta::AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows ordered character names" do
        expect(result.dig("data", "searchCharacters", "edges")).to match_array([
          {
            "node" => {
              "name" => character2.name,
              "favoriteCharactersCount" => 30
            }
          },
          {
            "node" => {
              "name" => character3.name,
              "favoriteCharactersCount" => 20
            }
          },
          {
            "node" => {
              "name" => character1.name,
              "favoriteCharactersCount" => 10
            }
          }
        ])
      end
    end
  end
end
