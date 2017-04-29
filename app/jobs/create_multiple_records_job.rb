# frozen_string_literal: true

class CreateMultipleRecordsJob < ApplicationJob
  queue_as :default

  def perform(user_id, episode_ids)
    user = User.find(user_id)
    episodes = Episode.published.where(id: episode_ids).order(:sort_number)

    return if episodes.blank?

    ActiveRecord::Base.transaction do
      multiple_record = user.multiple_records.create!(work: episodes.first.work)

      episodes.each do |episode|
        episode.records.create! do |c|
          c.user = user
          c.work = episode.work
          c.rating = 0
          c.multiple_record_id = multiple_record.id
        end

        latest_status = user.latest_statuses.find_by(work: episode.work)
        latest_status.append_episode(episode) if latest_status.present?
      end
    end
  end
end
