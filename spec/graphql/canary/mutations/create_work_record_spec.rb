# frozen_string_literal: true

xdescribe Canary::Mutations::CreateWorkRecord do
  let(:user) { create :registered_user }
  let(:work) { create :work }
  let(:token) { create(:oauth_access_token) }
  let(:context) { {viewer: user, doorkeeper_token: token, writable: true} }
  let(:work_id) { Canary::AnnictSchema.id_from_object(work, work.class) }
  let(:query) do
    <<~GRAPHQL
      mutation(
        $workId: ID!
        $comment: String
        $ratingOverall: Rating
        $ratingAnimation: Rating
        $ratingMusic: Rating
        $ratingStory: Rating
        $ratingCharacter: Rating
        $shareToTwitter: Boolean
      ) {
        createWorkRecord(
          input: {
            workId: $workId
            comment: $comment
            ratingOverall: $ratingOverall
            ratingAnimation: $ratingAnimation
            ratingMusic: $ratingMusic
            ratingStory: $ratingStory
            ratingCharacter: $ratingCharacter
            shareToTwitter: $shareToTwitter
          }
        ) {
          errors {
            message
          }

          record {
            databaseId
          }
        }
      }
    GRAPHQL
  end

  context "正常系" do
    let(:variables) do
      {
        workId: work_id,
        comment: "あうあう",
        ratingOverall: "GOOD",
        ratingAnimation: "GOOD",
        ratingMusic: "GOOD",
        ratingStory: "GOOD",
        ratingCharacter: "GOOD",
        shareToTwitter: false
      }
    end

    it "作成されたRecordオブジェクトが返ること" do
      expect(Record.count).to eq 0

      result = Canary::AnnictSchema.execute(query, variables: variables, context: context)

      expect(Record.count).to eq 1
      record = user.records.first

      expect(result["errors"]).to be_nil
      expect(result.dig("data", "createWorkRecord", "record", "databaseId")).to eq record.id
      expect(result.dig("data", "createWorkRecord", "errors")).to eq []
    end
  end

  context "異常系" do
    context "バリデーションエラーになったとき" do
      let(:variables) do
        {
          workId: work_id,
          comment: "a" * (1_048_596 + 1), # 文字数制限 (1,048,596文字) 以上の感想を書く
          ratingOverall: "GOOD",
          ratingAnimation: "GOOD",
          ratingMusic: "GOOD",
          ratingStory: "GOOD",
          ratingCharacter: "GOOD",
          shareToTwitter: false
        }
      end

      it "エラー内容が返ること" do
        result = Canary::AnnictSchema.execute(query, variables: variables, context: context)

        expect(result["errors"]).to be_nil
        expect(result.dig("data", "createWorkRecord", "record", "databaseId")).to be_nil

        errors = result.dig("data", "createWorkRecord", "errors")
        expect(errors.length).to eq 1
        expect(errors.first["message"]).to eq "感想は1048596文字以内で入力してください"
      end
    end
  end
end
