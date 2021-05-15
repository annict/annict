# frozen_string_literal: true

class UpdateStatusRepository < ApplicationRepository
  class RepositoryResult < Result
    attr_accessor :anime_entity
  end

  def execute(anime:, kind:)
    data = mutate(
      variables: {
        animeId: Canary::AnnictSchema.id_from_object(anime, Work),
        kind: Status.kind_v2_to_v3(kind)&.upcase&.to_s
      }
    )
    result = validate(data)

    if result.success?
      anime_node = data.dig("data", "updateStatus", "anime")
      result.anime_entity = AnimeEntity.from_node(anime_node)
    end

    result
  end

  private

  def result_class
    RepositoryResult
  end
end