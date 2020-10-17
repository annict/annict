# frozen_string_literal: true

module TrackPage
  class LibraryEntriesRepository < ApplicationRepository
    def execute
      result = query
      library_entry_nodes = result.to_h.dig("data", "viewer", "libraryEntries", "nodes")

      LibraryEntryEntity.from_nodes(library_entry_nodes)
    end
  end
end
