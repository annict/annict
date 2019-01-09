# frozen_string_literal: true

Mutations::UpdateReview = GraphQL::Relay::Mutation.define do
  name "UpdateReview"

  input_field :reviewId, !types.ID
  input_field :title, types.String
  input_field :body, !types.String
  WorkRecord::STATES.each do |state|
    input_field state.to_s.camelcase(:lower).to_sym, !Types::Enum::RatingState
  end
  input_field :shareTwitter, types.Boolean
  input_field :shareFacebook, types.Boolean

  return_field :review, ObjectTypes::Review

  resolve RescueFrom.new ->(_obj, inputs, ctx) {
    raise Annict::Errors::InvalidAPITokenScopeError unless ctx[:doorkeeper_token].writable?

    work_record = ctx[:viewer].work_records.published.find_by_graphql_id(inputs[:reviewId])

    work_record.title = inputs[:title]
    work_record.body = inputs[:body]
    WorkRecord::STATES.each do |state|
      work_record.send("#{state}=".to_sym, inputs[state.to_s.camelcase(:lower).to_sym]&.downcase)
    end
    work_record.modified_at = Time.now
    work_record.oauth_application = ctx[:doorkeeper_token].application
    work_record.detect_locale!(:body)

    ctx[:viewer].setting.attributes = {
      share_review_to_twitter: inputs[:shareTwitter] == true,
      share_review_to_facebook: inputs[:shareFacebook] == true
    }

    work_record.save!
    ctx[:viewer].setting.save!
    work_record.share_to_sns

    {
      review: work_record
    }
  }
end
