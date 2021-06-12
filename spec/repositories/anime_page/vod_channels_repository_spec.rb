# frozen_string_literal: true

describe V4::AnimePage::VodChannelsRepository, type: :repository do
  include V4::GraphqlRunnable

  let(:anime) { create :work }
  let(:channel) { Channel.published.first }
  let(:vod_channel) { Channel.published.with_vod.first }
  let(:anime_entity) do
    V4::AnimeEntity.new(
      programs: [
        V4::ProgramEntity.new(channel: V4::ChannelEntity.new(database_id: channel.id)),
        V4::ProgramEntity.new(channel: V4::ChannelEntity.new(database_id: vod_channel.id))
      ]
    )
  end

  context "正常系" do
    it "動画配信サービス情報が取得できること" do
      result = V4::AnimePage::VodChannelsRepository
        .new(graphql_client: graphql_client)
        .execute(anime_entity: anime_entity)
      vod_channel_entities = result.vod_channel_entities

      expect(vod_channel_entities.pluck(:database_id)).to contain_exactly(107, 165, 241, 243, 244, 260)
    end
  end
end
