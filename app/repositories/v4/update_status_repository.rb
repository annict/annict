# frozen_string_literal: true

module V4
  class UpdateStatusRepository < V4::ApplicationRepository
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
        result.anime_entity = V4::AnimeEntity.from_node(anime_node)
      end

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
