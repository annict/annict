# frozen_string_literal: true

module UserHome
  class FollowingActivityGroupsRepository < ApplicationRepository
    def execute(username:, pagination:)
      result = query(variables: {
        username: username,
        first: pagination.first,
        last: pagination.last,
        before: pagination.before,
        after: pagination.after
      })
      data = result.to_h.dig("data", "user", "followingActivityGroups")

      [ActivityGroupEntity.from_nodes(data["nodes"]), PageInfoEntity.from_node(data["pageInfo"])]
    end
  end
end
