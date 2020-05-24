# frozen_string_literal: true

module UserHome
  class FetchGlobalActivityGroupsRepository < ApplicationRepository
    include ActivityBuildable

    def fetch(cursor:)
      result = execute(variables: { cursor: cursor.presence || "" })
      data = result.to_h.dig("data", "activityGroups")

      {
        page_info: build_page_info(data["pageInfo"]),
        activity_groups: build_activity_groups(data["nodes"])
      }
    end
  end
end
