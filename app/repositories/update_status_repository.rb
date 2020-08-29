# frozen_string_literal: true

class UpdateStatusRepository < ApplicationRepository
  def execute(anime:, kind:)
    result = mutate(variables: {
      animeId: Canary::AnnictSchema.id_from_object(anime, Work),
      kind: Status.kind_v2_to_v3(kind)&.upcase&.to_s
    })

    if result.to_h["errors"]
      return [nil, MutationError.new(message: result.to_h["errors"][0]["message"])]
    end

    anime_node = result.dig("data", "updateStatus", "anime")

    [AnimeEntity.from_node(anime_node), nil]
  end
end
