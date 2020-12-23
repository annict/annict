# frozen_string_literal: true

module EpisodePage
  class EpisodeRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :episode_entity
    end

    def execute(database_id:)
      data = query(variables: { databaseId: database_id })
      episode_node = data.to_h.dig("data", "episode")

      result.episode_entity = EpisodeEntity.from_node(episode_node)

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
