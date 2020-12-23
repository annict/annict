# frozen_string_literal: true

module ProgramListModalContent
  class ProgramsRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :program_entities
    end

    def execute(anime_id:)
      data = query(variables: { animeId: anime_id })
      program_nodes = data.to_h.dig("data", "node", "programs", "nodes")

      result.program_entities = ProgramEntity.from_nodes(program_nodes)

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
