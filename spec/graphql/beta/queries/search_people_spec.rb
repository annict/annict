# typed: false
# frozen_string_literal: true

describe "GraphQL API Query" do
  describe "searchPeople" do
    let!(:person1) { create(:person, favorite_users_count: 10) }
    let!(:person2) { create(:person, favorite_users_count: 30) }
    let!(:person3) { create(:person, favorite_users_count: 20) }

    context "when `annictIds` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchPeople(annictIds: [#{person1.id}]) {
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

      it "shows person name" do
        expect(result.dig("data", "searchPeople", "edges")).to match_array([
          {
            "node" => {
              "name" => person1.name
            }
          }
        ])
      end
    end

    context "when `names` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchPeople(names: ["#{person3.name}"]) {
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

      it "shows person name" do
        expect(result.dig("data", "searchPeople", "edges")).to match_array([
          {
            "node" => {
              "name" => person3.name
            }
          }
        ])
      end
    end

    context "when `orderBy` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchPeople(orderBy: { field: FAVORITE_PEOPLE_COUNT, direction: DESC }) {
              edges {
                node {
                  name
                  favoritePeopleCount
                }
              }
            }
          }
        QUERY

        res = Beta::AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows ordered person names" do
        expect(result.dig("data", "searchPeople", "edges")).to match_array([
          {
            "node" => {
              "name" => person2.name,
              "favoritePeopleCount" => 30
            }
          },
          {
            "node" => {
              "name" => person3.name,
              "favoritePeopleCount" => 20
            }
          },
          {
            "node" => {
              "name" => person1.name,
              "favoritePeopleCount" => 10
            }
          }
        ])
      end
    end
  end
end
