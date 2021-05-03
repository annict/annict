# frozen_string_literal: true

module Canary
  module Mutations
    class SelectProgram < Canary::Mutations::Base
      argument :program_id, ID, required: true

      field :library_entry, Canary::Types::Objects::LibraryEntryType, null: false

      def resolve(program_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        program = Canary::AnnictSchema.object_from_id(program_id)

        library_entry = viewer.library_entries.where(work_id: program.work_id).first_or_create!
        library_entry.update!(program_id: program.id)

        {
          library_entry: library_entry
        }
      end
    end
  end
end
