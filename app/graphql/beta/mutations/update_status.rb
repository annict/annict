# frozen_string_literal: true

module Beta
  module Mutations
    class UpdateStatus < Beta::Mutations::Base
      include V4::GraphqlRunnable

      argument :work_id, ID, required: true
      argument :state, Beta::Types::Enums::StatusState, required: true

      field :work, Beta::Types::Objects::WorkType, null: true

      def resolve(work_id:, state:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

        work = Work.only_kept.find_by_graphql_id(work_id)

        state = case state
        when "NO_STATE" then "no_select"
        else
          state.downcase
        end

        UpdateStatusRepository.new(
          graphql_client: graphql_client(viewer: context[:viewer])
        ).execute(anime: work, kind: state)

        {
          work: work
        }
      end
    end
  end
end
