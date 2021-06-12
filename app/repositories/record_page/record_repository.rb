# frozen_string_literal: true

module RecordPage
  class RecordRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :user_entity, :record_entity
    end

    def execute(username:, record_database_id:)
      data = query(
        variables: {
          username: username,
          databaseId: record_database_id
        }
      )
      record_node = data.to_h.dig("data", "user", "record")

      result.record_entity = V4::RecordEntity.from_node(record_node)

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
