# frozen_string_literal: true

module TrackPage
  class LibraryEntriesRepository < ApplicationRepository
    class RepositoryResult < Result
      attr_accessor :library_entry_entities
    end

    def execute
      data = query
      library_entry_nodes = data.to_h.dig("data", "viewer", "libraryEntries", "nodes")

      result.library_entry_entities = LibraryEntryEntity.from_nodes(library_entry_nodes)

      result
    end

    private

    def result_class
      RepositoryResult
    end
  end
end
