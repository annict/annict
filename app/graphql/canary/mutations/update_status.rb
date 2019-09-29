# frozen_string_literal: true

module Canary
  module Mutations
    class UpdateStatus < Canary::Mutations::Base
      argument :work_id, ID, required: true
      argument :kind, Canary::Types::Enums::StatusKind, required: true

      field :work, Canary::Types::Objects::WorkType, null: true

      def resolve(work_id:, kind:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        work = Work.published.find_by_graphql_id(work_id)
        status = StatusService.new(context[:viewer], work)
        status.app = context[:application]
        status.via = context[:via]

        status.change!(Status.kind_v3_to_v2(kind.downcase).to_s)

        {
          work: work
        }
      end
    end
  end
end
