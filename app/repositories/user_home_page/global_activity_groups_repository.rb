# frozen_string_literal: true

module UserHomePage
  class GlobalActivityGroupsRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :activity_group_entities, :page_info_entity
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
      activity_groups_data = result.to_h.dig("data", "activityGroups")

      result.activity_group_entities = ActivityGroupEntity.from_nodes(activity_groups_data["nodes"])
      result.page_info_entity = PageInfoEntity.from_node(activity_groups_data["pageInfo"])

      result
    end
  end
end
