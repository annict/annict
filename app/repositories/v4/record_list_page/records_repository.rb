# frozen_string_literal: true

module V4::RecordListPage
  class RecordsRepository < V4::ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :record_entities, :page_info_entity
    end

    def execute(username:, pagination:, month: nil)
      data = query(
        variables: {
          username: username,
          month: month,
          first: pagination.first,
          last: pagination.last,
          before: pagination.before,
          after: pagination.after
        }
      )
      records_data = data.to_h.dig("data", "user", "records")

      result.record_entities = V4::RecordEntity.from_nodes(records_data["nodes"])
      result.page_info_entity = V4::PageInfoEntity.from_node(records_data["pageInfo"])

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
