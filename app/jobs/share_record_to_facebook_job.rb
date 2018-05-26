# frozen_string_literal: true

class ShareRecordToFacebookJob < ApplicationJob
  queue_as :default

  def perform(user_id, record_id)
    user = User.find(user_id)
    record = user.records.find(record_id)
    work_image = record.work.work_image

    image_url = if work_image.present? && Rails.env.production?
      work_image.decorate.image_url(:attachment, size: "600x315")
    else
      "https://annict.com/images/og_image.png"
    end

    FacebookService.new(user).share!(record, image_url)
  end
end
