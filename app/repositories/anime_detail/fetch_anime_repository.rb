# frozen_string_literal: true

module AnimeDetail
  class FetchAnimeRepository < ApplicationRepository
    def fetch(anime_id:)
      result = execute(variables: { databaseId: anime_id.to_i })
      data = result.to_h.dig("data", "animeList")
      anime_node = data["nodes"].first

      AnimeEntity.from_node(anime_node)
    end
  end
end
