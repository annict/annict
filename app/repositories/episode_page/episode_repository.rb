# frozen_string_literal: true

module EpisodePage
  class EpisodeRepository < ApplicationRepository
    def execute(database_id:)
      result = query(variables: { databaseId: database_id })
      episode_node = result.to_h.dig("data", "episode")

      EpisodeEntity.from_node(episode_node)
    end
  end
end
