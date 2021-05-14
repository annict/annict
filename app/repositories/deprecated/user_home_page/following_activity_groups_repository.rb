# frozen_string_literal: true

module UserHomePage
  class FollowingActivityGroupsRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :activity_group_entities, :page_info_entity
    end

    def execute(username:, pagination:)
      data = query(
        variables: {
          username: username,
          first: pagination.first,
          last: pagination.last,
          before: pagination.before,
          after: pagination.after
        }
      )
      activity_groups_data = data.to_h.dig("data", "user", "followingActivityGroups")

      result.activity_group_entities = ActivityGroupEntity.from_nodes(activity_groups_data["nodes"])
      result.page_info_entity = PageInfoEntity.from_node(activity_groups_data["pageInfo"])

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
