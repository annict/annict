# frozen_string_literal: true

describe "GraphQL API (Beta) Mutation" do
  describe "createReview" do
    let!(:user) { create(:user, :with_setting) }
    let!(:work) { create(:work) }
    let!(:token) { create(:oauth_access_token) }
    let!(:context) { {viewer: user, doorkeeper_token: token} }
    let!(:id) { Beta::AnnictSchema.id_from_object(work, work.class) }
    let!(:body) { "とてもよかった！" }
    let!(:result) do
      query_string = <<~GRAPHQL
        mutation {
          createReview(input: {
            workId: "#{id}",
            body: "#{body}",
            ratingAnimationState: GOOD,
            ratingCharacterState: BAD,
            ratingMusicState: GREAT,
            ratingOverallState: AVERAGE,
            ratingStoryState: GOOD,
          }) {
            review {
              ... on Review {
                id
                annictId
                body
                ratingAnimationState
                ratingCharacterState
                ratingMusicState
                ratingOverallState
                ratingStoryState
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

    it "create resource" do
      review = WorkRecord.last
      expect(result.dig("data", "createReview", "review", "annictId")).to eq(review.id)
      expect(result.dig("data", "createReview", "review", "body")).to eq(review.body)
      expect(result.dig("data", "createReview", "review", "ratingAnimationState")).to eq("GOOD")
      expect(result.dig("data", "createReview", "review", "ratingCharacterState")).to eq("BAD")
      expect(result.dig("data", "createReview", "review", "ratingMusicState")).to eq("GREAT")
      expect(result.dig("data", "createReview", "review", "ratingOverallState")).to eq("AVERAGE")
      expect(result.dig("data", "createReview", "review", "ratingAnimationState")).to eq("GOOD")
    end
  end
end
