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

        res = Beta::AnnictSchema.execute(query_string)
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

        res = Beta::AnnictSchema.execute(query_string)
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

        res = Beta::AnnictSchema.execute(query_string)
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

        res = Beta::AnnictSchema.execute(query_string)
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

    context "when `casts` are fetched" do
      let!(:cast1) { create(:cast, work: work1) }
      let!(:cast2) { create(:cast, work: work1) }
      let!(:cast3) { create(:cast, work: work2) }
      let(:result) do
        query_string = <<~QUERY
          query {
            searchWorks(orderBy: { field: WATCHERS_COUNT, direction: DESC }) {
              edges {
                node {
                  title
                  watchersCount
                  casts(orderBy: { field: CREATED_AT, direction: DESC }) {
                    edges {
                      node {
                        character {
                          name
                        }
                        person {
                          name
                        }
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

      it "shows ordered casts" do
        expect(result.dig("data", "searchWorks", "edges")).to match_array([
          {
            "node" => {
              "title" => work2.title,
              "watchersCount" => 30,
              "casts" => {
                "edges" => [
                  {
                    "node" => {
                      "character" => {
                        "name" => cast3.character.name
                      },
                      "person" => {
                        "name" => cast3.person.name
                      }
                    }
                  }
                ]
              }
            }
          },
          {
            "node" => {
              "title" => work3.title,
              "watchersCount" => 20,
              "casts" => {
                "edges" => []
              }
            }
          },
          {
            "node" => {
              "title" => work1.title,
              "watchersCount" => 10,
              "casts" => {
                "edges" => [
                  {
                    "node" => {
                      "character" => {
                        "name" => cast2.character.name
                      },
                      "person" => {
                        "name" => cast2.person.name
                      }
                    }
                  },
                  {
                    "node" => {
                      "character" => {
                        "name" => cast1.character.name
                      },
                      "person" => {
                        "name" => cast1.person.name
                      }
                    }
                  }
                ]
              }
            }
          }
        ])
      end
    end

    context "when `staffs` are fetched" do
      let(:organization) { create(:organization) }
      let!(:staff1) { create(:staff, work: work1) }
      let!(:staff2) { create(:staff, work: work1, resource: organization) }
      let!(:staff3) { create(:staff, work: work2) }
      let(:result) do
        query_string = <<~QUERY
          query {
            searchWorks(orderBy: { field: WATCHERS_COUNT, direction: DESC }) {
              edges {
                node {
                  title
                  watchersCount
                  staffs(orderBy: { field: CREATED_AT, direction: DESC }, first: 3) {
                    edges {
                      node {
                        resource {
                          ... on Person {
                            name
                          }
                          ... on Organization {
                            name
                          }
                        }
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

      it "shows ordered casts" do
        expect(result.dig("data", "searchWorks", "edges")).to match_array([
          {
            "node" => {
              "title" => work2.title,
              "watchersCount" => 30,
              "staffs" => {
                "edges" => [
                  {
                    "node" => {
                      "resource" => {
                        "name" => staff3.resource.name
                      }
                    }
                  }
                ]
              }
            }
          },
          {
            "node" => {
              "title" => work3.title,
              "watchersCount" => 20,
              "staffs" => {
                "edges" => []
              }
            }
          },
          {
            "node" => {
              "title" => work1.title,
              "watchersCount" => 10,
              "staffs" => {
                "edges" => [
                  {
                    "node" => {
                      "resource" => {
                        "name" => staff2.resource.name
                      }
                    }
                  },
                  {
                    "node" => {
                      "resource" => {
                        "name" => staff1.resource.name
                      }
                    }
                  }
                ]
              }
            }
          }
        ])
      end
    end

    context "when `reviews` are fetched" do
      let(:user) { create(:registered_user) }
      let!(:record) { create(:record, :on_work, user: user, work: work1, body: "Review~~~") }
      let(:result) do
        query_string = <<~QUERY
          query {
            searchWorks(annictIds: [#{work1.id}]) {
              nodes {
                reviews {
                  nodes {
                    body
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

      it "returns reviews" do
        expect(result.dig("data", "searchWorks", "nodes")).to match_array([
          {
            "reviews" => {
              "nodes" => [
                {
                  "body" => "Review~~~"
                }
              ]
            }
          }
        ])
      end
    end
  end
end
