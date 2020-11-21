# frozen_string_literal: true

class CheckProgramRepository < ApplicationRepository
  def execute(program_id:)
    result = mutate(variables: {
      programId: program_id
    })

    if result.to_h["errors"]
      return [nil, MutationError.new(message: result.to_h["errors"][0]["message"])]
    end

    library_entry_node = result.dig("data", "checkProgram", "libraryEntry")

    [LibraryEntryEntity.from_node(library_entry_node), nil]
  end
end
