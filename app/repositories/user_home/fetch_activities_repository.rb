# frozen_string_literal: true

module UserHome
  class FetchActivitiesRepository < ApplicationRepository
    def fetch(activity_group_id:, cursor:)
      result = execute(variables: { activityGroupId: activity_group_id, cursor: cursor.presence || "" })
      activity_group_node = result.to_h.dig("data", "node")

      ActivityGroupEntity.from_node(activity_group_node)
    end
  end
end
