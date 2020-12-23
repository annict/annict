# frozen_string_literal: true

module TrackAnimePage
  class AnimeRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :anime_entity, :page_info_entity
    end

    def execute(anime_id:, pagination:)
      data = query(
        variables: {
          databaseId: anime_id,
          first: pagination.first,
          last: pagination.last,
          before: pagination.before,
          after: pagination.after
        }
      )
      anime_node = data.to_h.dig("data", "anime")

      result.anime_entity = AnimeEntity.from_node(anime_node)
      result.page_info_entity = PageInfoEntity.from_node(anime_node.dig("episodes", "pageInfo"))

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
