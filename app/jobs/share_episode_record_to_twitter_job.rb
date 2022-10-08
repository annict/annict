# frozen_string_literal: true

class ShareEpisodeRecordToTwitterJob < ApplicationJob
  queue_as :default

  def perform(user_id, episode_record_id)
    user = User.only_kept.find(user_id)

    return unless user.share_record_to_twitter?

    episode_record = user.episode_records.only_kept.find(episode_record_id)

    Deprecated::TwitterService.new(user).share!(episode_record)
  end
end
