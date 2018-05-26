# frozen_string_literal: true

class ShareStatusToFacebookJob < ApplicationJob
  queue_as :default

  def perform(user_id, status_id)
    user = User.find(user_id)
    status= user.statuses.find(status_id)
    work_image = status.work.work_image

    image_url = if work_image.present? && Rails.env.production?
      work_image.decorate.image_url(:attachment, size: "600x315")
    else
      "https://annict.com/images/og_image.png"
    end

    FacebookService.new(user).share!(status, image_url)
  end
end
