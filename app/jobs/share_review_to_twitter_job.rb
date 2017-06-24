# frozen_string_literal: true

class ShareReviewToTwitterJob < ApplicationJob
  queue_as :default

  def perform(user_id, review_id)
    user = User.find(user_id)
    review = user.reviews.published.find(review_id)

    TwitterService.new(user).share_review!(review)
  end
end
