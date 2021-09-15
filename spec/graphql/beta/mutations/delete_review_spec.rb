# frozen_string_literal: true

describe "GraphQL API (Beta) Mutation" do
  describe "deleteReview" do
    let!(:user) { create :registered_user }
    let!(:work) { create :work }
    let!(:work_record) { create(:work_record) }
    let!(:record) { create :record, :for_work, user: user, work: work, recordable: work_record }
    let!(:activity_group) { create(:activity_group, user: user, itemable_type: "Record") }
    let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: record) }
    let!(:token) { create(:oauth_access_token) }
    let!(:context) { {viewer: user, doorkeeper_token: token, writable: true} }
    let!(:work_record_id) { Beta::AnnictSchema.id_from_object(work_record, work_record.class) }
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
      let(:variables) { {reviewId: work_record_id} }

      it "記録が削除されること" do
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1
        expect(Record.count).to eq 1
        expect(WorkRecord.count).to eq 1

        result = Beta::AnnictSchema.execute(query, variables: variables, context: context)
        expect(result["errors"]).to be_nil

        expect(ActivityGroup.count).to eq 0
        expect(Activity.count).to eq 0
        expect(Record.count).to eq 0
        expect(WorkRecord.count).to eq 0
      end
    end
  end
end
