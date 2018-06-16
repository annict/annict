# frozen_string_literal: true

class ShareWorkRecordToFacebookJob < ApplicationJob
  queue_as :default

  def perform(user_id, work_record_id)
    user = User.find(user_id)
    work_record = user.work_records.published.find(work_record_id)
    work_image = work_record.work.work_image

    image_url = if work_image.present? && Rails.env.production?
      work_image.decorate.image_url(:attachment, size: "600x315")
    else
      "https://annict.com/images/og_image.png"
    end

    FacebookService.new(user).share!(work_record, image_url)
  end
end
