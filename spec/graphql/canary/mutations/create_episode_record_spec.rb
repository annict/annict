# frozen_string_literal: true

xdescribe Canary::Mutations::CreateEpisodeRecord do
  let(:user) { create :registered_user }
  let(:episode) { create :episode }
  let(:work) { episode.work }
  let(:token) { create(:oauth_access_token) }
  let(:context) { {viewer: user, doorkeeper_token: token, writable: true} }
  let(:id) { Canary::AnnictSchema.id_from_object(episode, episode.class) }

  context "正常系" do
    let(:query) do
      <<~GRAPHQL
        mutation {
          createEpisodeRecord(input: {
            episodeId: "#{id}",
            comment: "にぱー",
            rating: GOOD
          }) {
            record {
              databaseId
            }
            errors {
              message
            }
          }
        }
      GRAPHQL
    end

    it "作成されたRecordオブジェクトが返ること" do
      expect(Record.count).to eq 0

      result = Canary::AnnictSchema.execute(query, context: context)

      expect(Record.count).to eq 1
      record = user.records.first

      expect(result["errors"]).to be_nil
      expect(result.dig("data", "createEpisodeRecord", "record", "databaseId")).to eq record.id
      expect(result.dig("data", "createEpisodeRecord", "errors")).to eq []
    end
  end

  context "異常系" do
    context "バリデーションエラーになったとき" do
      let(:query) do
        <<~GRAPHQL
          mutation {
            createEpisodeRecord(input: {
              episodeId: "#{id}",
              comment: "#{"a" * (1_048_596 + 1)}", # 文字数制限 (1,048,596文字) 以上の感想を書く
              rating: GOOD
            }) {
              record {
                databaseId
              }
              errors {
                message
              }
            }
          }
        GRAPHQL
      end

      it "エラー内容が返ること" do
        result = Canary::AnnictSchema.execute(query, context: context)

        expect(result["errors"]).to be_nil
        expect(result.dig("data", "createEpisodeRecord", "record", "databaseId")).to be_nil

        errors = result.dig("data", "createEpisodeRecord", "errors")
        expect(errors.length).to eq 1
        expect(errors.first["message"]).to eq "感想は1048596文字以内で入力してください"
      end
    end
  end
end
