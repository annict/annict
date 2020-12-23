# frozen_string_literal: true

module EpisodePage
  class MyRecordsRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :record_entities
    end

    def execute(episode_id:)
      data = query(variables: { episode_id: episode_id })
      record_nodes = data.to_h.dig("data", "viewer", "records", "nodes")

      result.record_entities = RecordEntity.from_nodes(record_nodes)

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
