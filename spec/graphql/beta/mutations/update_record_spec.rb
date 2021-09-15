# frozen_string_literal: true

describe Beta::Mutations::UpdateRecord do
  let!(:user) { create :registered_user }
  let!(:work) { create :work }
  let!(:episode) { create :episode, work: work }
  let!(:episode_record) { create(:episode_record) }
  let!(:record) { create :record, :on_episode, user: user, work: work, recordable: episode_record }
  let!(:token) { create(:oauth_access_token) }
  let!(:context) { {viewer: user, doorkeeper_token: token, writable: true} }
  let!(:episode_record_id) { Canary::AnnictSchema.id_from_object(episode_record, episode_record.class) }

  context "正常系" do
    let(:query) do
      <<~GRAPHQL
        mutation($recordId: ID!) {
          updateRecord(input: {
            recordId: $recordId,
            comment: "またーり",
            ratingState: GREAT
          }) {
            record {
              annictId
              comment
              ratingState
            }
          }
        }
      GRAPHQL
    end
    let(:variables) { {recordId: episode_record_id} }

    it "更新されたRecordデータが返ること" do
      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1
      expect(record.body).to eq "おもしろかった"
      expect(record.rating).to eq "good"

      record_cache_expired_at = user.record_cache_expired_at
      result = Beta::AnnictSchema.execute(query, variables: variables, context: context)

      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1

      expect(result["errors"]).to be_nil
      expect(result.dig("data", "updateRecord", "record", "comment")).to eq "またーり"
      expect(result.dig("data", "updateRecord", "record", "ratingState")).to eq "GREAT"
      expect(user.record_cache_expired_at).to_not eq record_cache_expired_at
    end
  end
end
