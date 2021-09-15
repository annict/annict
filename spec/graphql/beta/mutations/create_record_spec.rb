# frozen_string_literal: true

describe "GraphQL API (Beta) Mutation" do
  describe "createRecord" do
    let!(:episode) { create(:episode) }
    let!(:user) { create(:user, :with_setting) }
    let!(:token) { create(:oauth_access_token) }
    let!(:context) { {viewer: user, doorkeeper_token: token} }
    let!(:id) { Beta::AnnictSchema.id_from_object(episode, episode.class) }
    let!(:body) { "とてもよかった！" }
    let!(:result) do
      query_string = <<~GRAPHQL
        mutation {
          createRecord(input: {
            comment: "#{body}",
            episodeId: "#{id}",
            ratingState: GOOD
          }) {
            record {
              ... on Record {
                id
                annictId
                comment
                ratingState
              }
            }
          }
        }
      GRAPHQL

      res = Beta::AnnictSchema.execute(query_string, context: context)
      pp(res) if res["errors"]
      res
    end

    before do
      result
    end

    it "create episode record" do
      record = Record.last
      episode_record = record.episode_record
      expect(result.dig("data", "createRecord", "record", "annictId")).to eq(episode_record.id)
      expect(result.dig("data", "createRecord", "record", "comment")).to eq(record.body)
      expect(result.dig("data", "createRecord", "record", "ratingState")).to eq("GOOD")
    end
  end
end
