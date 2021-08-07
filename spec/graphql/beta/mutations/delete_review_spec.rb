# frozen_string_literal: true

describe "GraphQL API (Beta) Mutation" do
  describe "deleteReview" do
    let!(:user) { create :registered_user }
    let!(:anime) { create :anime }
    let!(:record) { create :record, user: user, anime: anime }
    let!(:anime_record) { create(:anime_record, user: user, record: record, anime: anime) }
    let!(:activity_group) { create(:activity_group, user: user, itemable_type: "AnimeRecord") }
    let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: anime_record) }
    let!(:token) { create(:oauth_access_token) }
    let!(:context) { {viewer: user, doorkeeper_token: token, writable: true} }
    let!(:anime_record_id) { Beta::AnnictSchema.id_from_object(anime_record, anime_record.class) }
    let!(:query) do
      <<~GRAPHQL
        mutation($reviewId: ID!) {
          deleteReview(input: { reviewId: $reviewId }) {
            work {
              title
            }
          }
        }
      GRAPHQL
    end

    context "正常系" do
      let(:variables) { {reviewId: anime_record_id} }

      it "記録が削除されること" do
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1
        expect(Record.count).to eq 1
        expect(AnimeRecord.count).to eq 1

        result = Beta::AnnictSchema.execute(query, variables: variables, context: context)
        expect(result["errors"]).to be_nil

        expect(ActivityGroup.count).to eq 0
        expect(Activity.count).to eq 0
        expect(Record.count).to eq 0
        expect(AnimeRecord.count).to eq 0
      end
    end
  end
end
