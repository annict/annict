# frozen_string_literal: true

describe AnimePage::VodChannelsRepository, type: :repository do
  include V4::GraphqlRunnable

  let(:anime) { create :work }
  let(:channel) { create :channel, :published }
  let(:vod_channel) { create :channel, :published, vod: true }
  let(:anime_entity) do
    AnimeEntity.new(
      programs: [
        ProgramEntity.new(channel: ChannelEntity.new(database_id: channel.id)),
        ProgramEntity.new(channel: ChannelEntity.new(database_id: vod_channel.id))
      ]
    )
  end

  context "正常系" do
    it "動画配信サービス情報が取得できること" do
      result = AnimePage::VodChannelsRepository
        .new(graphql_client: graphql_client)
        .execute(anime_entity: anime_entity)

      # VODなチャンネルと非VODなチャンネルが1つずつ存在するはず
      expect(Channel.count).to eq 2

      vod_channel_entities = result.vod_channel_entities
      expect(vod_channel_entities.length).to eq 1

      vod_channel_entity = vod_channel_entities.first
      # VODなチャンネルだけ取得できるはず
      expect(vod_channel_entity.database_id).to eq vod_channel.id
    end
  end
end
