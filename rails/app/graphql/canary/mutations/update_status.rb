# typed: false
# frozen_string_literal: true

module Canary
  module Mutations
    class UpdateStatus < Canary::Mutations::Base
      argument :work_id, ID, required: true
      argument :kind, Canary::Types::Enums::StatusKind, required: true

      field :work, Canary::Types::Objects::WorkType, null: true

      def resolve(work_id:, kind:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        work = Work.only_kept.find_by_graphql_id(work_id)

        form = Forms::StatusForm.new(work: work, kind: kind)

        if form.invalid?
          raise GraphQL::ExecutionError, "status update failed"
        end

        Updaters::StatusUpdater.new(user: viewer, form: form).call

        {
          work: work
        }
      end
    end
  end
end
