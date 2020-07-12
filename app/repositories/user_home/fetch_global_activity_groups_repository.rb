# frozen_string_literal: true

module UserHome
  class FetchGlobalActivityGroupsRepository < ApplicationRepository
    def fetch(pagination:)
      result = execute(variables: {
        first: pagination.first,
        last: pagination.last,
        before: pagination.before,
        after: pagination.after
      })
      data = result.to_h.dig("data", "activityGroups")

      [ActivityGroupEntity.from_nodes(data["nodes"]), PageInfoEntity.from_node(data["pageInfo"])]
    end
  end
end
