# frozen_string_literal: true

class ShareEpisodeRecordToTwitterJob < ApplicationJob
  queue_as :default

  def perform(user_id, episode_record_id)
    user = User.find(user_id)
    episode_record = user.episode_records.without_deleted.find(episode_record_id)

    TwitterService.new(user).share!(episode_record)
  end
end
