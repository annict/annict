# frozen_string_literal: true

module ProgramListModalContent
  class ProgramsRepository < ApplicationRepository
    def execute(anime_id:)
      result = query(variables: { animeId: anime_id })
      program_nodes = result.to_h.dig("data", "node", "programs", "nodes")

      ProgramEntity.from_nodes(program_nodes)
    end
  end
end
