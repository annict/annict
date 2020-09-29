# frozen_string_literal: true

module EpisodePage
  class EpisodeRepository < ApplicationRepository
    def execute(episode_id:)
      result = query(variables: { databaseId: episode_id })
      episode_node = result.to_h.dig("data", "episode")

      EpisodeEntity.from_node(episode_node)
    end
  end
end
