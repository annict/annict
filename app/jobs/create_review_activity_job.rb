# frozen_string_literal: true

class CreateReviewActivityJob < ApplicationJob
  queue_as :default

  def perform(user_id, review_id)
    user = User.find(user_id)
    review = user.reviews.published.find(review_id)

    Activity.create! do |a|
      a.user = user
      a.recipient = review.work
      a.trackable = review
      a.action = "create_review"
      a.work = review.work
      a.review = review
    end
  end
end
