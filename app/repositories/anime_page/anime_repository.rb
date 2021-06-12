# frozen_string_literal: true

module V4::AnimePage
  class AnimeRepository < V4::ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :anime_entity
    end

    def execute(anime_id:)
      data = query(variables: {databaseId: anime_id.to_i})
      anime_node = data.to_h.dig("data", "animeList", "nodes").first

      result.anime_entity = V4::AnimeEntity.from_node(anime_node)

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
