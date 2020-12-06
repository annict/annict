# frozen_string_literal: true

module TrackPage
  class SlotsRepository < ApplicationRepository
    def execute(pagination:)
      result = query(
        variables: {
          first: pagination.first,
          last: pagination.last,
          before: pagination.before,
          after: pagination.after
        }
      )
      slot_nodes = result.to_h.dig("data", "viewer", "slots", "nodes")
      page_info_node = result.to_h.dig("data", "viewer", "slots", "pageInfo")

      [SlotEntity.from_nodes(slot_nodes), PageInfoEntity.from_node(page_info_node)]
    end
  end
end
