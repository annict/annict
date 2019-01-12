# frozen_string_literal: true

describe "GraphQL API Query" do
  describe "searchWorks" do
    let!(:work1) { create(:work, :with_current_season, watchers_count: 10) }
    let!(:work2) { create(:work, :with_next_season, watchers_count: 30) }
    let!(:work3) { create(:work, :with_prev_season, watchers_count: 20) }

    context "when `annictIds` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchWorks(annictIds: [#{work1.id}]) {
              edges {
                node {
                  title
                }
              }
            }
          }
        QUERY

        res = AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows work title" do
        expect(result.dig("data", "searchWorks", "edges")).to match_array([
          {
            "node" => {
              "title" => work1.title
            }
          }
        ])
      end
    end

    context "when `seasons` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchWorks(seasons: ["#{work2.season.slug}"]) {
              edges {
                node {
                  title
                }
              }
            }
          }
        QUERY

        res = AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows work title" do
        expect(result.dig("data", "searchWorks", "edges")).to match_array([
          {
            "node" => {
              "title" => work2.title
            }
          }
        ])
      end
    end

    context "when `titles` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchWorks(titles: ["#{work3.title}"]) {
              edges {
                node {
                  title
                }
              }
            }
          }
        QUERY

        res = AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows work title" do
        expect(result.dig("data", "searchWorks", "edges")).to match_array([
          {
            "node" => {
              "title" => work3.title
            }
          }
        ])
      end
    end

    context "when `orderBy` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            searchWorks(orderBy: { field: WATCHERS_COUNT, direction: DESC }) {
              edges {
                node {
                  title
                  watchersCount
                }
              }
            }
          }
        QUERY

        res = AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows ordered work titles" do
        expect(result.dig("data", "searchWorks", "edges")).to match_array([
          {
            "node" => {
              "title" => work2.title,
              "watchersCount" => 30
            }
          },
          {
            "node" => {
              "title" => work3.title,
              "watchersCount" => 20
            }
          },
          {
            "node" => {
              "title" => work1.title,
              "watchersCount" => 10
            }
          }
        ])
      end
    end
  end
end
