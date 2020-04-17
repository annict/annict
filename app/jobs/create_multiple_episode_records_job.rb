# frozen_string_literal: true

class CreateMultipleEpisodeRecordsJob < ApplicationJob
  queue_as :default

  def perform(user_id, episode_ids)
    user = User.find(user_id)
    episodes = Episode.only_kept.where(id: episode_ids).order(:sort_number)

    return if episodes.blank?

    ActiveRecord::Base.transaction do
      multiple_episode_record = user.multiple_episode_records.create!(work: episodes.first.work)

      episodes.each do |episode|
        episode.episode_records.create! do |c|
          c.user = user
          c.work = episode.work
          c.record = user.records.create!(work: episode.work)
          c.rating = 0
          c.multiple_episode_record = multiple_episode_record
        end

        library_entry = user.library_entries.find_by(work: episode.work)
        library_entry.append_episode!(episode) if library_entry.present?
      end
    end
  end
end
