# frozen_string_literal: true

class BulkCreateEpisodeRecordsRepository < ApplicationRepository
  def execute(form:)
    result = mutate(variables: { episodeIds: form.episode_ids })

    if result.to_h["errors"]
      return [nil, MutationError.new(message: result.to_h["errors"][0]["message"])]
    end

    bulk_operation_node = result.dig("data", "bulkCreateEpisodeRecords", "bulkOperation")

    [BulkOperationEntity.from_node(bulk_operation_node), nil]
  end
end
