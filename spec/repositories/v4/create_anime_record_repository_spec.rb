# frozen_string_literal: true

describe V4::CreateAnimeRecordRepository, type: :repository do
  include V4::GraphqlRunnable

  let(:user) { create :registered_user }
  let(:anime) { create :work }
  let(:anime_id) { Canary::AnnictSchema.id_from_object(anime, anime.class) }
  let(:form) do
    Forms::AnimeRecordForm.new(
      anime: anime,
      rating_overall: "GOOD",
      rating_animation: "GOOD",
      rating_music: "GOOD",
      rating_story: "GOOD",
      rating_character: "GOOD",
      comment: comment,
      share_to_twitter: false
    )
  end

  context "正常系" do
    let(:comment) { "すごく面白かった。" }

    it "RecordEntityが返ること" do
      expect(Record.count).to eq 0

      result = V4::CreateAnimeRecordRepository.new(
        graphql_client: graphql_client(viewer: user)
      ).execute(form: form)

      expect(Record.count).to eq 1
      record = user.records.first

      expect(result.errors).to be_empty
      expect(result.record_entity.database_id).to eq record.id
    end
  end

  context "異常系" do
    context "バリデーションエラーになるとき" do
      let(:comment) { "a" * (1_048_596 + 1) } # 文字数制限 (1,048,596文字) 以上の感想

      it "エラー内容が返ること" do
        result = V4::CreateAnimeRecordRepository.new(
          graphql_client: graphql_client(viewer: user)
        ).execute(form: form)

        expect(result.record_entity).to be_nil
        expect(result.errors.length).to eq 1
        expect(result.errors.first.message).to eq "感想は1048596文字以内で入力してください"
      end
    end
  end
end
