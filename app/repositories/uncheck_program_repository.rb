# frozen_string_literal: true

class UncheckProgramRepository < ApplicationRepository
  def execute(anime_id:)
    result = mutate(variables: {
      animeId: anime_id
    })

    if result.to_h["errors"]
      return [nil, MutationError.new(message: result.to_h["errors"][0]["message"])]
    end

    library_entry_node = result.dig("data", "uncheckProgram", "libraryEntry")

    [LibraryEntryEntity.from_node(library_entry_node), nil]
  end
end
