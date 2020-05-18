# frozen_string_literal: true

module ProfileDetail
  class FetchUserActivityGroupsRepository < ApplicationRepository
    include UserHome::ActivityBuildable

    def fetch(username:, cursor:)
      result = execute(variables: { username: username, cursor: cursor.presence || "" })
      data = result.to_h.dig("data", "user", "activityGroups")

      {
        page_info: build_page_info(data["pageInfo"]),
        activity_groups: build_activity_groups(data["nodes"])
      }
    end
  end
end
