# frozen_string_literal: true

Mutations::DeleteRecord = GraphQL::Relay::Mutation.define do
  name "DeleteRecord"

  input_field :recordId, !types.ID

  return_field :episode, ObjectTypes::Episode

  resolve RescueFrom.new ->(_obj, inputs, ctx) {
    raise Annict::Errors::InvalidAPITokenScopeError unless ctx[:doorkeeper_token].writable?

    record = ctx[:viewer].records.find_by_graphql_id(inputs[:recordId])
    record.destroy

    {
      episode: record.episode
    }
  }
end
