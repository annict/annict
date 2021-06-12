# frozen_string_literal: true

module AnimeRecordListPage
  class AnimeRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :anime_entity
    end

    def execute(database_id:)
      data = query(variables: {databaseId: database_id})
      anime_node = data.to_h.dig("data", "anime")

      result.anime_entity = V4::AnimeEntity.from_node(anime_node)

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
