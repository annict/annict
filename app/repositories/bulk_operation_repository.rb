# frozen_string_literal: true

class BulkOperationRepository < ApplicationRepository
  def execute(job_id:)
    result = query(variables: { jobId: job_id })
    bulk_operation_node = result.to_h.dig("data", "bulkOperation")

    if bulk_operation_node
      BulkOperationEntity.from_node(bulk_operation_node)
    end
  end
end
