# frozen_string_literal: true

module ProfileDetail
  class UserActivityGroupsRepository < ApplicationRepository
    def execute(username:, pagination:)
      result = query(variables: {
        username: username,
        first: pagination.first,
        last: pagination.last,
        before: pagination.before,
        after: pagination.after
      })
      activity_groups_data = result.to_h.dig("data", "user", "activityGroups")

      [ActivityGroupEntity.from_nodes(activity_groups_data["nodes"]), PageInfoEntity.from_node(activity_groups_data["pageInfo"])]
    end
  end
end
