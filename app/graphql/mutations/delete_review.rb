# frozen_string_literal: true

Mutations::DeleteReview = GraphQL::Relay::Mutation.define do
  name "DeleteReview"

  input_field :reviewId, !types.ID

  return_field :work, ObjectTypes::Work

  resolve RescueFrom.new ->(_obj, inputs, ctx) {
    raise Annict::Errors::InvalidAPITokenScopeError unless ctx[:doorkeeper_token].writable?

    work_record = ctx[:viewer].work_records.published.find_by_graphql_id(inputs[:reviewId])
    work_record.record.destroy

    {
      work: work_record.work
    }
  }
end
