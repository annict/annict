# frozen_string_literal: true

module RecordList
  class RecordsRepository < ApplicationRepository
    def execute(username:, pagination:, month: nil)
      result = query(
        variables: {
          username: username,
          month: month,
          first: pagination.first,
          last: pagination.last,
          before: pagination.before,
          after: pagination.after
        }
      )
      records_data = result.to_h.dig("data", "user", "records")

      [RecordEntity.from_nodes(records_data["nodes"]), PageInfoEntity.from_node(records_data["pageInfo"])]
    end
  end
end
