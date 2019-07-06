# frozen_string_literal: true

module Canary
  module Mutations
    class StatusUpdate < Canary::Mutations::Base
      argument :work_id, ID, required: true
      argument :state, Canary::Types::Enums::StatusState, required: true

      field :work, Canary::Types::Objects::WorkType, null: true

      def resolve(work_id:, state:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        work = Work.published.find_by_graphql_id(work_id)
        status = StatusService.new(context[:viewer], work)
        status.app = context[:application]
        status.via = context[:via]

        state = case state
        when "NO_STATE" then "no_select"
        else
          state.downcase
        end

        status.change!(state)

        {
          work: work
        }
      end
    end
  end
end
