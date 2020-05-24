# frozen_string_literal: true

module UserHome
  class FetchActivitiesRepository < ApplicationRepository
    include ActivityBuildable

    def fetch(activity_group_id:, cursor:)
      result = execute(variables: { activityGroupId: activity_group_id, cursor: cursor.presence || "" })
      node = result.to_h.dig("data", "node")

      ActivityGroupEntity.new(
        itemable_type: node["itemableType"].downcase,
        itemables: build_itemables(node.dig("activities", "nodes"), build_user(node["user"]))
      )
    end
  end
end
