# frozen_string_literal: true

class UpdateStatusRepository < ApplicationRepository
  def create(work:, kind:)
    result = execute(variables: {
      workId: Canary::AnnictSchema.id_from_object(work, Work),
      kind: Status.kind_v2_to_v3(kind)&.upcase&.to_s
    })

    if result.to_h["errors"]
      return [nil, MutationError.new(message: result.to_h["errors"][0]["message"])]
    end

    work_node = result.dig("data", "updateStatus", "work")

    [WorkEntity.from_node(work_node), nil]
  end
end
