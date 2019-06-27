# frozen_string_literal: true

module Mutations
  class UpdateStatus < Mutations::Base
    argument :work_id, ID, required: false
    argument :work_annict_id, Integer, required: false
    argument :state, Types::Enums::StatusState, required: true

    field :work, Types::Objects::WorkType, null: true

    def resolve(work_id: nil, work_annict_id: nil, state:)
      # raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

      work = work_id ? Work.published.find_by_graphql_id(work_id) : Work.published.find(work_annict_id)
      status = StatusService.new(context[:viewer], work)
      # status.app = context[:doorkeeper_token].application
      # status.ga_client = context[:ga_client]
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
