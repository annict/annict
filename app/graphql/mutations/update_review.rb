# frozen_string_literal: true

Mutations::UpdateReview = GraphQL::Relay::Mutation.define do
  name "UpdateReview"

  input_field :reviewId, !types.ID
  input_field :title, !types.String
  input_field :body, !types.String
  Review::STATES.each do |state|
    input_field state.to_s.camelcase(:lower).to_sym, !EnumTypes::RatingState
  end
  input_field :shareTwitter, types.Boolean
  input_field :shareFacebook, types.Boolean

  return_field :review, ObjectTypes::Review

  resolve RescueFrom.new ->(_obj, inputs, ctx) {
    raise Annict::Errors::InvalidAPITokenScopeError unless ctx[:doorkeeper_token].writable?

    review = ctx[:viewer].reviews.published.find_by_graphql_id(inputs[:reviewId])

    review.title = inputs[:title]
    review.body = inputs[:body]
    Review::STATES.each do |state|
      review.send("#{state}=".to_sym, inputs[state.to_s.camelcase(:lower).to_sym]&.downcase)
    end
    review.modified_at = Time.now
    review.oauth_application = ctx[:doorkeeper_token].application

    ctx[:viewer].setting.attributes = {
      share_review_to_twitter: inputs[:shareTwitter] == true,
      share_review_to_facebook: inputs[:shareFacebook] == true
    }

    review.save!
    ctx[:viewer].setting.save!
    review.share_to_sns

    {
      review: review
    }
  }
end
