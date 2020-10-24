# frozen_string_literal: true

class AddReactionRepository < ApplicationRepository
  def execute(reactable:, content:)
    result = mutate(
      variables: {
        reactableId: Canary::AnnictSchema.id_from_object(reactable, reactable.class),
        content: content
      }
    )

    if result.to_h["errors"]
      return [nil, MutationError.new(message: result.to_h["errors"][0]["message"])]
    end

    [nil, nil]
  end
end
