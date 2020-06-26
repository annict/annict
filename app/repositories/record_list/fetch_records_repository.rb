# frozen_string_literal: true

module RecordList
  class FetchRecordsRepository < ApplicationRepository
    def fetch(username:, before:, after:, per:, month: nil)
      result = execute(
        variables: {
          username: username,
          month: month,
          first: per,
          last: per,
          before: before,
          after: after
        }
      )
      records = result.to_h.dig("data", "user", "records")

      [PageInfoEntity.from_node(records["pageInfo"]), build_records(records["nodes"])]
    end

    private

    def build_records(record_nodes)
      record_nodes.map do |record_node|
        RecordEntity.from_node(record_node)
      end
    end
  end
end
