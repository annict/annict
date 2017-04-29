# frozen_string_literal: true

class FacebookShareJob < ApplicationJob
  queue_as :default

  def perform(user_id, record_id)
    user = User.find(user_id)
    record = Checkin.find(record_id)
    work_image = record.work.work_image

    source = if work_image.present? && Rails.env.production?
      work_image.decorate.image_url(:attachment, size: "600x315")
    else
      "https://annict.com/images/og_image.png"
    end

    FacebookService.new(user).share!(record, source)
  end
end
