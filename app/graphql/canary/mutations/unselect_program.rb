# frozen_string_literal: true

module Canary
  module Mutations
    class UnselectProgram < Canary::Mutations::Base
      argument :anime_id, ID, required: true

      field :library_entry, Canary::Types::Objects::LibraryEntryType, null: true

      def resolve(anime_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        anime = Canary::AnnictSchema.object_from_id(anime_id)

        library_entry = viewer.library_entries.find_by(work_id: anime.id)

        unless library_entry
          return {}
        end

        library_entry.update!(program_id: nil)

        {
          library_entry: library_entry
        }
      end
    end
  end
end
