# frozen_string_literal: true

module Canary
  module Mutations
    class UnselectProgram < Canary::Mutations::Base
      argument :work_id, ID, required: true

      field :library_entry, Canary::Types::Objects::LibraryEntryType, null: true

      def resolve(work_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        work = Canary::AnnictSchema.object_from_id(work_id)

        library_entry = viewer.library_entries.find_by(work_id: work.id)

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
