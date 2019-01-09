# frozen_string_literal: true

Mutations::UpdateRecord = GraphQL::Relay::Mutation.define do
  name "UpdateRecord"

  input_field :recordId, !types.ID
  input_field :comment, types.String
  input_field :ratingState, Types::Enum::RatingState
  input_field :shareTwitter, types.Boolean
  input_field :shareFacebook, types.Boolean

  return_field :record, ObjectTypes::Record

  resolve RescueFrom.new ->(_obj, inputs, ctx) {
    raise Annict::Errors::InvalidAPITokenScopeError unless ctx[:doorkeeper_token].writable?

    record = ctx[:viewer].episode_records.published.find_by_graphql_id(inputs[:recordId])

    record.rating_state = inputs[:ratingState]&.downcase
    record.modify_comment = record.comment != inputs[:comment]
    record.comment = inputs[:comment]
    record.shared_twitter = inputs[:shareTwitter] == true
    record.shared_facebook = inputs[:shareFacebook] == true
    record.oauth_application = ctx[:doorkeeper_token].application
    record.detect_locale!(:comment)

    record.save!
    record.update_share_record_status
    record.share_to_sns

    {
      record: record
    }
  }
end
