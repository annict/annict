# frozen_string_literal: true

module Canary
  module Mutations
    class UnlikeWorkRecord < Canary::Mutations::Base
      argument :work_record_id, ID, required: true

      field :work_record, Canary::Types::Objects::WorkRecordType, null: false

      def resolve(work_record_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        work_record = WorkRecord.only_kept.find_by_graphql_id(work_record_id)
        context[:viewer].unlike(work_record)

        {
          work_record: work_record
        }
      end
    end
  end
end
