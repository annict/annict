# frozen_string_literal: true

module Mutations
  class UpdateStatus < Mutations::Base
    argument :work_id, ID, required: true
    argument :state, Types::Enums::StatusState, required: true

    field :work, Types::Objects::WorkType, null: true

    def resolve(work_id:, state:)
      raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

      work = Work.only_kept.find_by_graphql_id(work_id)

      state = case state
      when "NO_STATE" then "no_select"
      else
        state.downcase
      end

      ChangeStatusService.new(user: context[:viewer], work: work).call(status_kind: state)

      {
        work: work
      }
    end
  end
end
