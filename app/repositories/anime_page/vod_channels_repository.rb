# frozen_string_literal: true

module AnimePage
  class VodChannelsRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :vod_channel_entities
    end

    def execute(anime_entity:)
      data = query
      channel_nodes = data.to_h.dig("data", "channels", "nodes")

      result.vod_channel_entities = V4::VodChannelEntity.from_nodes(channel_nodes, anime_entity: anime_entity)

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
