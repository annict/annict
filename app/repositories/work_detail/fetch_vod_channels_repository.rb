# frozen_string_literal: true

module WorkDetail
  class FetchVodChannelsRepository < ApplicationRepository
    def fetch(work_entity:)
      result = execute
      channel_nodes = result.to_h.dig("data", "channels", "nodes")

      VodChannelEntity.from_nodes(channel_nodes, work_entity: work_entity)
    end
  end
end
