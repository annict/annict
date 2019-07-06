# frozen_string_literal: true

module Canary
  module Mutations
    class WorkRecordDelete < Canary::Mutations::Base
      argument :work_record_id, ID, required: true

      field :work, Canary::Types::Objects::WorkType, null: true

      def resolve(work_record_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        work_record = context[:viewer].work_records.published.find_by_graphql_id(work_record_id)
        work_record.record.destroy

        {
          work: work_record.work
        }
      end
    end
  end
end
