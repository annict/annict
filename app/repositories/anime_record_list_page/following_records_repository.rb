# frozen_string_literal: true

module AnimeRecordListPage
  class FollowingRecordsRepository < ApplicationRepository
    def execute(anime_id:)
      result = query(variables: { anime_id: anime_id })
      record_nodes = result.to_h.dig("data", "node", "records", "nodes")

      RecordEntity.from_nodes(record_nodes)
    end
  end
end
