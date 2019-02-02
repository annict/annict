# frozen_string_literal: true

module Mutations
  class UpdateStatus < Mutations::Base
    argument :work_id, ID, required: true
    argument :state, Types::Enums::StatusState, required: true

    field :work, Types::Objects::WorkType, null: true

    def resolve(work_id:, state:)
      raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

      work = Work.published.find_by_graphql_id(work_id)
      status = StatusService.new(context[:viewer], work)
      status.app = context[:doorkeeper_token].application
      status.ga_client = context[:ga_client]
      status.keen_client = context[:keen_client]
      status.via = "graphql_api"

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
