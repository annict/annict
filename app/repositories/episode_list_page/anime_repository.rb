# frozen_string_literal: true

module EpisodeListPage
  class AnimeRepository < ApplicationRepository
    def execute(database_id:, pagination:)
      result = query(variables: {
        databaseId: database_id,
        first: pagination.first,
        last: pagination.last,
        before: pagination.before,
        after: pagination.after
      })
      anime_node = result.to_h.dig("data", "anime")

      [AnimeEntity.from_node(anime_node), PageInfoEntity.from_node(anime_node.dig("episodes", "pageInfo"))]
    end
  end
end
