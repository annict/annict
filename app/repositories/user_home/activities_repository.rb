# frozen_string_literal: true

module UserHome
  class ActivitiesRepository < ApplicationRepository
    def execute(activity_group_id:, cursor:)
      result = query(variables: { activityGroupId: activity_group_id, cursor: cursor.presence || "" })
      activity_group_node = result.to_h.dig("data", "node")

      ActivityGroupEntity.from_node(activity_group_node)
    end
  end
end
