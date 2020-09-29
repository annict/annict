# frozen_string_literal: true

module AnimePage
  class VodChannelsRepository < ApplicationRepository
    def execute(anime_entity:)
      result = query
      channel_nodes = result.to_h.dig("data", "channels", "nodes")

      VodChannelEntity.from_nodes(channel_nodes, anime_entity: anime_entity)
    end
  end
end
