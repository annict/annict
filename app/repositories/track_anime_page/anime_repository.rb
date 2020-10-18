# frozen_string_literal: true

module TrackAnimePage
  class AnimeRepository < ApplicationRepository
    def execute(anime_id:)
      result = query(variables: { databaseId: anime_id })
      anime_node = result.to_h.dig("data", "anime")

      AnimeEntity.from_node(anime_node)
    end
  end
end
