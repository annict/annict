# frozen_string_literal: true

describe AnimeRecordListPage::AnimeRepository, type: :repository do
  include V4::GraphqlRunnable

  let(:anime) { create :work }

  context "正常系" do
    it "アニメ情報が取得できること" do
      result = AnimeRecordListPage::AnimeRepository.new(graphql_client: graphql_client).execute(database_id: anime.id)
      anime_entity = result.anime_entity

      expect(anime_entity.database_id).to eq anime.id
      expect(anime_entity.title).to eq anime.title
    end
  end
end
