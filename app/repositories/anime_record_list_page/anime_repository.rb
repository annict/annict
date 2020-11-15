# frozen_string_literal: true

module AnimeRecordListPage
  class AnimeRepository < ApplicationRepository
    def execute(database_id:)
      result = query(variables: { databaseId: database_id })
      anime_node = result.to_h.dig("data", "anime")

      AnimeEntity.from_node(anime_node)
    end
  end
end
