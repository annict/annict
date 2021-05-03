# frozen_string_literal: true

module UserHomePage
  class ActivitiesRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :activity_group_entity
    end

    def execute(activity_group_id:, cursor:)
      data = query(variables: {activityGroupId: activity_group_id, cursor: cursor.presence || ""})
      activity_group_node = data.to_h.dig("data", "node")

      result.activity_group_entity = ActivityGroupEntity.from_node(activity_group_node)

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
