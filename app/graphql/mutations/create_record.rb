# frozen_string_literal: true

Mutations::CreateRecord = GraphQL::Relay::Mutation.define do
  name "CreateRecord"

  input_field :episodeId, !types.ID
  input_field :comment, types.String
  input_field :ratingState, EnumTypes::RatingState
  input_field :shareTwitter, types.Boolean
  input_field :shareFacebook, types.Boolean

  return_field :record, ObjectTypes::Record

  resolve RescueFrom.new ->(_obj, inputs, ctx) {
    raise Annict::Errors::InvalidAPITokenScopeError unless ctx[:doorkeeper_token].writable?

    episode = Episode.published.find_by_graphql_id(inputs[:episodeId])

    record = episode.episode_records.new do |r|
      r.rating_state = inputs[:ratingState]&.downcase
      r.comment = inputs[:comment]
      r.shared_twitter = inputs[:shareTwitter] == true
      r.shared_facebook = inputs[:shareFacebook] == true
      r.oauth_application = ctx[:doorkeeper_token].application
    end

    service = NewEpisodeRecordService.new(ctx[:viewer], record)
    service.ga_client = ctx[:ga_client]
    service.timber = ctx[:timber]
    service.app = ctx[:doorkeeper_token].application
    service.via = "graphql_api"

    service.save!

    {
      record: service.episode_record
    }
  }
end
