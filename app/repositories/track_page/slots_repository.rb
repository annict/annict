# frozen_string_literal: true

module TrackPage
  class SlotsRepository < ApplicationRepository
    def execute
      result = query
      slot_nodes = result.to_h.dig("data", "viewer", "slots", "nodes")

      SlotEntity.from_nodes(slot_nodes)
    end
  end
end
