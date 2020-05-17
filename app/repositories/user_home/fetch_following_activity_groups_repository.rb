# frozen_string_literal: true

module UserHome
  class FetchFollowingActivityGroupsRepository < ApplicationRepository
    include ActivityBuildable

    def fetch(username:, cursor:)
      result = execute(variables: { username: username, cursor: cursor.presence || "" })
      data = result.to_h.dig("data", "user", "followingActivityGroups")

      {
        page_info: build_page_info(data["pageInfo"]),
        activity_groups: build_activity_groups(data["nodes"])
      }
    end
  end
end
