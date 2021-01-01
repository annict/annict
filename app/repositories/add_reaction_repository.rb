# frozen_string_literal: true

class AddReactionRepository < ApplicationRepository
  class RepositoryResult < Result; end

  def execute(reactable:, content:)
    data = mutate(
      variables: {
        reactableId: Canary::AnnictSchema.id_from_object(reactable, reactable.class),
        content: content
      }
    )

    validate(data)
  end
end
