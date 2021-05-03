# frozen_string_literal: true

describe Canary::Mutations::UpdateAnimeRecord do
  let(:user) { create :registered_user }
  let!(:record) { create(:record, user: user) }
  let!(:anime_record) { create(:work_record, user: user, record: record) }
  let(:token) { create(:oauth_access_token) }
  let(:context) { {viewer: user, doorkeeper_token: token, writable: true} }
  let(:record_id) { GraphQL::Schema::UniqueWithinType.encode(record.class.name, record.id) }

  context "正常系" do
    let(:query) do
      <<~GRAPHQL
        mutation {
          updateAnimeRecord(input: {
            recordId: "#{record_id}",
            ratingOverall: GREAT,
            ratingAnimation: GREAT,
            ratingMusic: GREAT,
            ratingStory: GREAT,
            ratingCharacter: GREAT,
            comment: "またーり"
          }) {
            record {
              databaseId
              comment

              recordable {
                ... on AnimeRecord {
                  ratingOverall
                  ratingAnimation
                  ratingMusic
                  ratingStory
                  ratingCharacter
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
      expect(WorkRecord.count).to eq 1
      expect(anime_record.body).to eq "おもしろかった"
      expect(anime_record.rating_overall_state).to be_nil
      expect(anime_record.rating_animation_state).to be_nil
      expect(anime_record.rating_music_state).to be_nil
      expect(anime_record.rating_story_state).to be_nil
      expect(anime_record.rating_character_state).to be_nil

      record_cache_expired_at = user.record_cache_expired_at
      result = Canary::AnnictSchema.execute(query, context: context)

      expect(Record.count).to eq 1
      expect(WorkRecord.count).to eq 1

      expect(result["errors"]).to be_nil
      expect(result.dig("data", "updateAnimeRecord", "record", "comment")).to eq "またーり"
      expect(result.dig("data", "updateAnimeRecord", "record", "recordable", "ratingOverall")).to eq "GREAT"
      expect(result.dig("data", "updateAnimeRecord", "record", "recordable", "ratingAnimation")).to eq "GREAT"
      expect(result.dig("data", "updateAnimeRecord", "record", "recordable", "ratingMusic")).to eq "GREAT"
      expect(result.dig("data", "updateAnimeRecord", "record", "recordable", "ratingStory")).to eq "GREAT"
      expect(result.dig("data", "updateAnimeRecord", "record", "recordable", "ratingCharacter")).to eq "GREAT"
      expect(result.dig("data", "updateAnimeRecord", "errors")).to eq []
      expect(user.record_cache_expired_at).to_not eq record_cache_expired_at
    end
  end
end
