# frozen_string_literal: true

module TrackAnimePage
  class AnimeRepository < ApplicationRepository
    def execute(anime_id:, pagination:)
      result = query(
        variables: {
          databaseId: anime_id,
          first: pagination.first,
          last: pagination.last,
          before: pagination.before,
          after: pagination.after
        }
      )
      anime_node = result.to_h.dig("data", "anime")

      [AnimeEntity.from_node(anime_node), PageInfoEntity.from_node(anime_node.dig("episodes", "pageInfo"))]
    end
  end
end
