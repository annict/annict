# frozen_string_literal: true

class ShareWorkRecordToFacebookJob < ApplicationJob
  queue_as :default

  def perform(user_id, review_id)
    user = User.find(user_id)
    review = user.reviews.published.find(review_id)
    work_image = review.work.work_image

    image_url = if work_image.present? && Rails.env.production?
      work_image.decorate.image_url(:attachment, size: "600x315")
    else
      "https://annict.com/images/og_image.png"
    end

    FacebookService.new(user).share!(review, image_url)
  end
end
