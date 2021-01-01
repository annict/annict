# frozen_string_literal: true

class BulkCreateEpisodeRecordsRepository < ApplicationRepository
  class RepositoryResult < Result
    attr_accessor :bulk_operation_entity
  end

  def execute(form:)
    data = mutate(variables: { episodeIds: form.episode_ids })
    result = validate(data)

    if result.success?
      bulk_operation_node = data.dig("data", "bulkCreateEpisodeRecords", "bulkOperation")
      result.bulk_operation_entity = BulkOperationEntity.from_node(bulk_operation_node)
    end

    result
  end
end
