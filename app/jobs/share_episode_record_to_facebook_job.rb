# frozen_string_literal: true

class ShareEpisodeRecordToFacebookJob < ApplicationJob
  queue_as :default

  def perform(user_id, episode_record_id)
    user = User.find(user_id)
    episode_record = user.episode_records.published.find(episode_record_id)
    work_image = episode_record.work.work_image

    image_url = if work_image.present? && Rails.env.production?
      work_image.decorate.image_url(:attachment, size: "600x315")
    else
      "https://annict.com/images/og_image.png"
    end

    FacebookService.new(user).share!(episode_record, image_url)
  end
end
