# frozen_string_literal: true

module AnimePage
  class AnimeRepository < ApplicationRepository
    def execute(anime_id:)
      result = query(variables: { databaseId: anime_id.to_i })
      data = result.to_h.dig("data", "animeList")
      anime_node = data["nodes"].first

      AnimeEntity.from_node(anime_node)
    end
  end
end
