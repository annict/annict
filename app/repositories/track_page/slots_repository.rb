# frozen_string_literal: true

module TrackPage
  class SlotsRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :slot_entities, :page_info_entity
    end

    def execute(pagination:)
      data = query(
        variables: {
          first: pagination.first,
          last: pagination.last,
          before: pagination.before,
          after: pagination.after
        }
      )
      slot_nodes = data.to_h.dig("data", "viewer", "slots", "nodes")
      page_info_node = data.to_h.dig("data", "viewer", "slots", "pageInfo")

      result.slot_entities = SlotEntity.from_nodes(slot_nodes)
      result.page_info_entity = PageInfoEntity.from_node(page_info_node)

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
