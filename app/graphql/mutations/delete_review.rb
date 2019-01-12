# frozen_string_literal: true

module Mutations
  class DeleteReview < Mutations::Base
    argument :review_id, ID, required: true

    field :work, Types::Objects::WorkType, null: false

    def resolve(review_id:)
      raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

      work_record = context[:viewer].work_records.published.find_by_graphql_id(review_id)
      work_record.record.destroy

      {
        work: work_record.work
      }
    end
  end
end
