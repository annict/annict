# frozen_string_literal: true

class BulkOperationRepository < ApplicationRepository
  class RepositoryResult < Result
    attr_accessor :bulk_operation_entity
  end

  def execute(job_id:)
    data = query(variables: { jobId: job_id })
    bulk_operation_node = data.to_h.dig("data", "bulkOperation")

    if bulk_operation_node
      result.bulk_operation_entity = BulkOperationEntity.from_node(bulk_operation_node)
    end

    result
  end
end
