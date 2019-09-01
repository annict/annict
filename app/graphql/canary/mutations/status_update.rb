# frozen_string_literal: true

module Canary
  module Mutations
    class StatusUpdate < Canary::Mutations::Base
      argument :work_id, ID, required: true
      argument :kind, Canary::Types::Enums::StatusKind, required: true

      field :work, Canary::Types::Objects::WorkType, null: true

      def resolve(work_id:, kind:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        work = Work.published.find_by_graphql_id(work_id)
        status = StatusService.new(context[:viewer], work)
        status.app = context[:application]
        status.via = context[:via]

        kind = case kind
        when "NO_STATE" then "no_select"
        else
          kind.downcase
        end

        status.change!(kind)

        {
          work: work
        }
      end
    end
  end
end
