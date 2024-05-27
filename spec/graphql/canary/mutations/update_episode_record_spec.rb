# frozen_string_literal: true

xdescribe Canary::Mutations::UpdateEpisodeRecord do
  let!(:user) { create :registered_user }
  let!(:work) { create :work }
  let!(:episode) { create :episode, work: work }
  let!(:record) { create :record, user: user, work: work }
  let!(:episode_record) { create(:episode_record, user: user, record: record, work: work, episode: episode, rating: nil) }
  let!(:token) { create(:oauth_access_token) }
  let!(:context) { {viewer: user, doorkeeper_token: token, writable: true} }
  let!(:record_id) { Canary::AnnictSchema.id_from_object(record, record.class) }

  context "正常系" do
    let(:query) do
      <<~GRAPHQL
        mutation($recordId: ID!) {
          updateEpisodeRecord(input: {
            recordId: $recordId,
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
    let(:variables) { {recordId: record_id} }

    it "更新されたRecordデータが返ること" do
      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1
      expect(episode_record.body).to eq "おもしろかった"
      expect(episode_record.rating_state).to be_nil

      record_cache_expired_at = user.record_cache_expired_at
      result = Canary::AnnictSchema.execute(query, variables: variables, context: context)

      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1

      expect(result["errors"]).to be_nil
      expect(result.dig("data", "updateEpisodeRecord", "errors")).to be_empty
      expect(result.dig("data", "updateEpisodeRecord", "record", "comment")).to eq "またーり"
      expect(result.dig("data", "updateEpisodeRecord", "record", "recordable", "rating")).to eq "GREAT"
      expect(user.record_cache_expired_at).not_to eq record_cache_expired_at
    end
  end
end
