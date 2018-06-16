# frozen_string_literal: true

Mutations::DeleteRecord = GraphQL::Relay::Mutation.define do
  name "DeleteRecord"

  input_field :recordId, !types.ID

  return_field :episode, ObjectTypes::Episode

  resolve RescueFrom.new ->(_obj, inputs, ctx) {
    raise Annict::Errors::InvalidAPITokenScopeError unless ctx[:doorkeeper_token].writable?

    episode_record = ctx[:viewer].episode_records.published.find_by_graphql_id(inputs[:recordId])
    episode_record.record.destroy

    {
      episode: episode_record.episode
    }
  }
end
