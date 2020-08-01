# frozen_string_literal: true

module WorkDetail
  class FetchWorkRepository < ApplicationRepository
    def fetch(work_id:)
      result = execute(variables: { databaseId: work_id.to_i })
      data = result.to_h.dig("data", "works")
      work_node = data["nodes"].first

      AnimeEntity.from_node(work_node)
    end
  end
end
