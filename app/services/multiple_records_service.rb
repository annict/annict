# frozen_string_literal: true

class MultipleRecordsService
  def initialize(user)
    @user = user
  end

  def save!(episode_ids)
    episodes = Episode.where(id: episode_ids).order(:sort_number)

    return if episodes.blank?

    ActiveRecord::Base.transaction do
      multiple_record = @user.multiple_records.create!(work: episodes.first.work)

      episodes.each do |episode|
        episode.checkins.create! do |c|
          c.user = @user
          c.work = episode.work
          c.rating = 0
          c.multiple_record_id = multiple_record.id
        end

        update_latest_status(episode)
      end
    end
  end

  private

  def update_latest_status(episode)
    latest_status = @user.latest_statuses.find_by(work: episode.work)
    latest_status.append_episode(episode) if latest_status.present?
  end
end
