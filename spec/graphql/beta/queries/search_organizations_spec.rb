# frozen_string_literal: true

describe "GraphQL API Query" do
  describe "searchOrganizations" do
    let!(:organization1) { create(:organization, favorite_users_count: 10) }
    let!(:organization2) { create(:organization, favorite_users_count: 30) }
    let!(:organization3) { create(:organization, favorite_users_count: 20) }

    context "when `annictIds` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchOrganizations(annictIds: [#{organization1.id}]) {
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

      it "shows organization name" do
        expect(result.dig("data", "searchOrganizations", "edges")).to match_array([
          {
            "node" => {
              "name" => organization1.name
            }
          }
        ])
      end
    end

    context "when `names` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchOrganizations(names: ["#{organization3.name}"]) {
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

      it "shows organization name" do
        expect(result.dig("data", "searchOrganizations", "edges")).to match_array([
          {
            "node" => {
              "name" => organization3.name
            }
          }
        ])
      end
    end

    context "when `orderBy` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchOrganizations(orderBy: { field: FAVORITE_ORGANIZATIONS_COUNT, direction: DESC }) {
              edges {
                node {
                  name
                  favoriteOrganizationsCount
                }
              }
            }
          }
        QUERY

        res = Beta::AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows ordered organization names" do
        expect(result.dig("data", "searchOrganizations", "edges")).to match_array([
          {
            "node" => {
              "name" => organization2.name,
              "favoriteOrganizationsCount" => 30
            }
          },
          {
            "node" => {
              "name" => organization3.name,
              "favoriteOrganizationsCount" => 20
            }
          },
          {
            "node" => {
              "name" => organization1.name,
              "favoriteOrganizationsCount" => 10
            }
          }
        ])
      end
    end
  end
end
