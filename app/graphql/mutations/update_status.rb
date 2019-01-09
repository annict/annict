# frozen_string_literal: true

Mutations::UpdateStatus = GraphQL::Relay::Mutation.define do
  name "UpdateStatus"

  input_field :workId, !types.ID
  input_field :state, !Types::Enum::StatusState

  return_field :work, ObjectTypes::Work

  resolve RescueFrom.new ->(_obj, inputs, ctx) {
    raise Annict::Errors::InvalidAPITokenScopeError unless ctx[:doorkeeper_token].writable?

    work = Work.published.find_by_graphql_id(inputs[:workId])
    status = StatusService.new(ctx[:viewer], work)
    status.app = ctx[:doorkeeper_token].application
    status.ga_client = ctx[:ga_client]
    status.logentries = ctx[:logentries]
    status.via = "graphql_api"

    state = case inputs[:state]
    when "NO_STATE" then "no_select"
    else
      inputs[:state].downcase
    end

    status.change!(state)

    {
      work: work
    }
  }
end
