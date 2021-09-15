# frozen_string_literal: true

describe Beta::Mutations::UpdateReview do
  let(:user) { create :registered_user }
  let!(:work_record) { create(:work_record) }
  let!(:record) { create(:record, :on_work, user: user, recordable: work_record, rating: nil) }
  let(:token) { create(:oauth_access_token) }
  let(:context) { {viewer: user, doorkeeper_token: token, writable: true} }
  let(:work_record_id) { Canary::AnnictSchema.id_from_object(work_record, work_record.class) }

  context "正常系" do
    let(:query) do
      <<~GRAPHQL
        mutation($reviewId: ID!) {
          updateReview(input: {
            reviewId: $reviewId,
            ratingOverallState: GREAT,
            ratingAnimationState: GREAT,
            ratingMusicState: GREAT,
            ratingStoryState: GREAT,
            ratingCharacterState: GREAT,
            body: "またーり"
          }) {
            review {
              annictId
              body
              ratingOverallState
              ratingAnimationState
              ratingMusicState
              ratingStoryState
              ratingCharacterState
            }
          }
        }
      GRAPHQL
    end
    let(:variables) { {reviewId: work_record_id} }

    it "更新されたRecordデータが返ること" do
      expect(Record.count).to eq 1
      expect(WorkRecord.count).to eq 1
      expect(record.body).to eq "おもしろかった"
      expect(record.rating).to be_nil
      expect(record.animation_rating).to be_nil
      expect(record.music_rating).to be_nil
      expect(record.story_rating).to be_nil
      expect(record.character_rating).to be_nil

      record_cache_expired_at = user.record_cache_expired_at
      result = Beta::AnnictSchema.execute(query, variables: variables, context: context)

      expect(Record.count).to eq 1
      expect(WorkRecord.count).to eq 1

      expect(result["errors"]).to be_nil
      expect(result.dig("data", "updateReview", "review", "body")).to eq "またーり"
      expect(result.dig("data", "updateReview", "review", "ratingOverallState")).to eq "GREAT"
      expect(result.dig("data", "updateReview", "review", "ratingAnimationState")).to eq "GREAT"
      expect(result.dig("data", "updateReview", "review", "ratingMusicState")).to eq "GREAT"
      expect(result.dig("data", "updateReview", "review", "ratingStoryState")).to eq "GREAT"
      expect(result.dig("data", "updateReview", "review", "ratingCharacterState")).to eq "GREAT"
      expect(user.record_cache_expired_at).to_not eq record_cache_expired_at
    end
  end
end
