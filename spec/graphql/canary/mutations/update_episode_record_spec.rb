# frozen_string_literal: true

describe Canary::Mutations::UpdateEpisodeRecord do
  let(:user) { create :registered_user }
  let(:episode_record) { create(:episode_record, user: user, rating: nil) }
  let!(:record) { episode_record.record }
  let(:token) { create(:oauth_access_token) }
  let(:context) { { viewer: user, doorkeeper_token: token, writable: true } }
  let(:record_id) { GraphQL::Schema::UniqueWithinType.encode(record.class.name, record.id) }

  context "正常系" do
    let(:query) do
      <<~GRAPHQL
        mutation {
          updateEpisodeRecord(input: {
            recordId: "#{record_id}",
            comment: "またーり",
            rating: GREAT
          }) {
            record {
              databaseId
              comment

              recordable {
                ... on EpisodeRecord {
                  rating
                }
              }
            }

            errors {
              message
            }
          }
        }
      GRAPHQL
    end

    it "更新されたRecordデータが返ること" do
      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1
      expect(episode_record.body).to eq "おもしろかった"
      expect(episode_record.rating_state).to be_nil

      record_cache_expired_at = user.record_cache_expired_at
      result = Canary::AnnictSchema.execute(query, context: context)

      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1

      expect(result["errors"]).to be_nil
      expect(result.dig("data", "updateEpisodeRecord", "record", "comment")).to eq "またーり"
      expect(result.dig("data", "updateEpisodeRecord", "record", "recordable", "rating")).to eq "GREAT"
      expect(result.dig("data", "updateEpisodeRecord", "errors")).to eq []
      expect(user.record_cache_expired_at).to_not eq record_cache_expired_at
    end
  end
end
