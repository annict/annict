# frozen_string_literal: true

Mutations::CreateRecord = GraphQL::Relay::Mutation.define do
  name "CreateRecord"

  input_field :episodeId, !types.ID
  input_field :comment, types.String
  input_field :rating, EnumTypes::RatingState
  input_field :shareTwitter, types.Boolean
  input_field :shareFacebook, types.Boolean

  return_field :record, ObjectTypes::Record

  resolve RescueFrom.new ->(_obj, inputs, ctx) {
    raise Annict::Errors::InvalidAPITokenScopeError unless ctx[:doorkeeper_token].writable?

    episode = Episode.published.find_by_graphql_id(inputs[:episodeId])
    rating = case inputs[:rating]
    when "GOOD" then 4.0
    when "BAD" then 2.0
    end

    record = episode.records.new do |r|
      r.rating = rating
      r.comment = inputs[:comment]
      r.shared_twitter = inputs[:shareTwitter] == true
      r.shared_facebook = inputs[:shareFacebook] == true
      r.oauth_application = ctx[:doorkeeper_token].application
    end

    service = NewRecordService.new(ctx[:viewer], record)
    service.keen_client = ctx[:keen_client]
    service.ga_client = ctx[:ga_client]
    service.app = ctx[:doorkeeper_token].application

    service.save!

    {
      record: service.record
    }
  }
end
