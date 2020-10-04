# frozen_string_literal: true

module EpisodePage
  class MyRecordsRepository < ApplicationRepository
    def execute(episode_id:)
      result = query(variables: { episode_id: episode_id })
      record_nodes = result.to_h.dig("data", "viewer", "records", "nodes")

      RecordEntity.from_nodes(record_nodes)
    end
  end
end
