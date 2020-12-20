# frozen_string_literal: true

describe CreateEpisodeRecordRepository, type: :repository do
  include V4::GraphqlRunnable

  let(:user) { create :registered_user }
  let(:episode) { create :episode }
  let(:episode_id) { Canary::AnnictSchema.id_from_object(episode, episode.class) }

  context "正常系" do
    it "RecordEntityが返ること" do
      expect(Record.count).to eq 0

      form = EpisodeRecordForm.new(
        episode_id: episode_id,
        comment: "すごく面白かった。",
        rating: "GREAT"
      )
      result = CreateEpisodeRecordRepository.new(
        graphql_client: graphql_client(viewer: user)
      ).execute(form: form)

      expect(Record.count).to eq 1
      record = user.records.first

      expect(result.errors).to be_empty
      expect(result.record_entity.database_id).to eq record.id
    end
  end

  context "異常系" do
    it "エラー内容が返ること" do
      form = EpisodeRecordForm.new(
        episode_id: episode_id,
        comment: "a" * (1_048_596 + 1), # 文字数制限 (1,048,596文字) 以上の感想を書く
        rating: "GREAT"
      )
      result = CreateEpisodeRecordRepository.new(
        graphql_client: graphql_client(viewer: user)
      ).execute(form: form)

      expect(result.record_entity).to be_nil
      expect(result.errors.length).to eq 1
      expect(result.errors.first.message).to eq "感想は1048596文字以内で入力してください"
    end
  end
end
