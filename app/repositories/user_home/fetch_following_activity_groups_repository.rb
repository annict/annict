# frozen_string_literal: true

module UserHome
  class FetchFollowingActivityGroupsRepository < ApplicationRepository
    include Itemable

    def fetch(username:, cursor:)
      result = execute(variables: { username: username, cursor: cursor.presence || "" })
      data = result.to_h.dig("data", "user", "followingActivityGroups")

      {
        page_info: PageInfoEntity.new(
          end_cursor: data.dig("pageInfo", "endCursor"),
          has_next_page: data.dig("pageInfo", "hasNextPage")
        ),
        activity_groups: data["nodes"].map do |node|
          ActivityGroupEntity.new(
            id: node["id"],
            itemable_type: node["itemableType"].downcase,
            single: node["single"],
            activities_count: node["activitiesCount"],
            created_at: node["createdAt"],
            user: build_user(node["user"]),
            itemables: build_itemables(node.dig("activities", "nodes"), build_user(node["user"])),
            activities_page_info: PageInfoEntity.new(
              end_cursor: node.dig("activities", "pageInfo", "endCursor")
            )
          )
        end
      }
    end
  end
end
