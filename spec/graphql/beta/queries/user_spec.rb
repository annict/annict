# frozen_string_literal: true

describe "GraphQL API Query" do
  describe "user" do
    let!(:user) { create(:user) }

    context "when `username` argument is specified" do
      let(:result) do
        query_string = <<~QUERY
          query {
            user(username: "#{user.username}") {
              username
            }
          }
        QUERY

        res = Beta::AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows user's username" do
        expect(result.dig("data", "user")).to eq(
          "username" => user.username
        )
      end
    end

    describe "`activities` field" do
      context "when `orderBy` argument is specified" do
        let!(:activity1) { create(:create_episode_record_activity, user: user) }
        let!(:activity2) { create(:create_episode_record_activity, user: user) }
        let!(:activity3) { create(:create_episode_record_activity, user: user) }
        let(:result) do
          query_string = <<~QUERY
            query {
              user(username: "#{user.username}") {
                username
                activities(orderBy: { field: CREATED_AT, direction: DESC }) {
                  edges {
                    annictId
                    action
                    node {
                      __typename
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

        it "shows ordered user's activities" do
          expect(result.dig("data", "user")).to eq(
            "username" => user.username,
            "activities" => {
              "edges" => [
                {
                  "annictId" => activity3.id,
                  "action" => "CREATE",
                  "node" => {
                    "__typename" => "Record"
                  }
                },
                {
                  "annictId" => activity2.id,
                  "action" => "CREATE",
                  "node" => {
                    "__typename" => "Record"
                  }
                },
                {
                  "annictId" => activity1.id,
                  "action" => "CREATE",
                  "node" => {
                    "__typename" => "Record"
                  }
                }
              ]
            }
          )
        end
      end
    end

    describe "`works` field" do
      context "when `state` argument is specified" do
        let!(:status1) { create(:status, user: user, kind: :watching) }
        let!(:status2) { create(:status, user: user, kind: :watched) }
        let!(:work1) { status1.work }
        let!(:work2) { status2.work }
        let!(:library_entry1) { create(:library_entry, user: user, work: work1, status: status1) }
        let!(:library_entry2) { create(:library_entry, user: user, work: work2, status: status2) }
        let(:result) do
          query_string = <<~QUERY
            query {
              user(username: "#{user.username}") {
                username
                works(state: WATCHING) {
                  edges {
                    node {
                      title
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

        it "shows user's activities" do
          expect(result.dig("data", "user")).to eq(
            "username" => user.username,
            "works" => {
              "edges" => [
                {
                  "node" => {
                    "title" => status1.work.title
                  }
                }
              ]
            }
          )
        end
      end
    end

    describe "`libraryEntries` フィールド" do
      let!(:status1) { create(:status, user: user, kind: :watching) }
      let!(:status2) { create(:status, user: user, kind: :watched) }
      let!(:work1) { status1.work }
      let!(:work2) { status2.work }
      let!(:library_entry1) { create(:library_entry, user: user, work: work1, status: status1) }
      let!(:library_entry2) { create(:library_entry, user: user, work: work2, status: status2) }
      let!(:library_entry3) { create(:library_entry, user: user, status: nil) } # ステータス未指定
      let!(:work3) { library_entry3.work }

      context "`states` が指定されているとき" do
        let(:query_string) do
          <<~QUERY
            query {
              user(username: "#{user.username}") {
                libraryEntries(states: [WATCHING]) {
                  nodes {
                    work {
                      title
                    }
                  }
                }
              }
            }
          QUERY
        end

        it "指定したステータスのLibraryEntryが返ること" do
          result = Beta::AnnictSchema.execute(query_string)
          expect(result["errors"]).to be_nil

          expect(result.dig("data", "user")).to eq(
            "libraryEntries" => {
              "nodes" => [
                {
                  "work" => {
                    "title" => work1.title
                  }
                }
              ]
            }
          )
        end
      end

      context "`states` が指定されていないとき" do
        let(:query_string) do
          <<~QUERY
            query {
              user(username: "#{user.username}") {
                libraryEntries {
                  nodes {
                    work {
                      title
                    }
                  }
                }
              }
            }
          QUERY
        end

        it "全てのLibraryEntryが返ること" do
          result = Beta::AnnictSchema.execute(query_string)
          expect(result["errors"]).to be_nil

          expect(result.dig("data", "user")).to eq(
            "libraryEntries" => {
              "nodes" => [
                {
                  "work" => {
                    "title" => work1.title
                  }
                },
                {
                  "work" => {
                    "title" => work2.title
                  }
                },
                {
                  "work" => {
                    "title" => work3.title
                  }
                }
              ]
            }
          )
        end
      end
    end

    describe "`avatar_url` field" do
      let(:result) do
        query_string = <<~QUERY
          query {
            user(username: "#{user.username}") {
              username
              avatarUrl
            }
          }
        QUERY

        res = Beta::AnnictSchema.execute(query_string)
        pp(res) if res["errors"]
        res
      end

      it "shows user's avatar image URL" do
        expect(result.dig("data", "user")).to eq(
          "username" => user.username,
          "avatarUrl" => "#{ENV.fetch("ANNICT_URL")}/dummy_image"
        )
      end
    end
  end
end
