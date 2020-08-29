# frozen_string_literal: true

module AnimeDetail
  class FetchVodChannelsRepository < ApplicationRepository
    def fetch(anime_entity:)
      result = execute
      channel_nodes = result.to_h.dig("data", "channels", "nodes")

      VodChannelEntity.from_nodes(channel_nodes, anime_entity: anime_entity)
    end
  end
end
