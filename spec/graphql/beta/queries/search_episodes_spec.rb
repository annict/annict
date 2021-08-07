# frozen_string_literal: true

describe "GraphQL API Query" do
  describe "searchEpisodes" do
    let!(:work) { create(:anime, :with_current_season) }
    let!(:episode1) { create(:episode, anime: work, sort_number: 1) }
    let!(:episode2) { create(:episode, anime: work, sort_number: 3) }
    let!(:episode3) { create(:episode, anime: work, sort_number: 2) }

    context "when `annictIds` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchEpisodes(annictIds: [#{episode1.id}]) {
              edges {
                node {
                  annictId
                  title
                }
              }
            }
          }
        QUERY

        res = Beta::AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows episode" do
        expect(result.dig("data", "searchEpisodes", "edges")).to match_array(
          [
            {
              "node" => {
                "annictId" => episode1.id,
                "title" => episode1.title
              }
            }
          ]
        )
      end
    end

    context "when `orderBy` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchEpisodes(orderBy: { field: SORT_NUMBER, direction: DESC }) {
              edges {
                node {
                  annictId
                  title
                  sortNumber
                }
              }
            }
          }
        QUERY

        res = Beta::AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows ordered episodes" do
        expect(result.dig("data", "searchEpisodes", "edges")).to match_array(
          [
            {
              "node" => {
                "annictId" => episode2.id,
                "title" => episode2.title,
                "sortNumber" => 3
              }
            },
            {
              "node" => {
                "annictId" => episode3.id,
                "title" => episode3.title,
                "sortNumber" => 2
              }
            },
            {
              "node" => {
                "annictId" => episode1.id,
                "title" => episode1.title,
                "sortNumber" => 1
              }
            }
          ]
        )
      end
    end

    context "when `recodes` are fetched" do
      let!(:record) { create(:episode_record, episode: episode1) }
      let(:result) do
        query_string = <<~QUERY
          query {
            searchEpisodes(orderBy: { field: SORT_NUMBER, direction: ASC }, first: 1) {
              edges {
                node {
                  annictId
                  records {
                    edges {
                      node {
                        annictId
                        comment
                      }
                    }
                  }
                }
              }
            }
          }
        QUERY

        res = Beta::AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows records" do
        expect(result.dig("data", "searchEpisodes", "edges")).to match_array(
          [
            {
              "node" => {
                "annictId" => episode1.id,
                "records" => {
                  "edges" => [
                    {
                      "node" => {
                        "annictId" => record.id,
                        "comment" => record.body
                      }
                    }
                  ]
                }
              }
            }
          ]
        )
      end
    end
  end
end
