# typed: false
# frozen_string_literal: true

class BulkCreateEpisodeRecordsJob < ApplicationJob
  queue_as :default

  def perform(user_id, episode_ids)
    user = User.only_kept.find(user_id)
    episodes = Episode.only_kept.where(id: episode_ids).order(:sort_number)

    return if episodes.blank?

    ActiveRecord::Base.transaction do
      episodes.each do |episode|
        EpisodeRecordCreator.new(user: user, episode: episode).call
      end
    end
  end
end
