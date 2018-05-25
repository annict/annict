# frozen_string_literal: true

class CreateEpisodeRecordActivityJob < ApplicationJob
  queue_as :default

  def perform(user_id, episode_record_id)
    user = User.find(user_id)
    episode_record = user.episode_records.find(episode_record_id)

    Activity.create! do |a|
      a.user = user
      a.recipient = episode_record.episode
      a.trackable = episode_record
      a.action = "create_episode_record"
      a.work = episode_record.work
      a.episode = episode_record.episode
      a.episode_record = episode_record
    end
  end
end
