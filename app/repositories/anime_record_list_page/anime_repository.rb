# frozen_string_literal: true

module AnimeRecordListPage
  class AnimeRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :anime_entity
    end

    def execute(database_id:)
      result = query(variables: { databaseId: database_id })
      anime_node = result.to_h.dig("data", "anime")

      result.anime_entity = AnimeEntity.from_node(anime_node)

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
