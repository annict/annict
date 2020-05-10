# frozen_string_literal: true

class CreateEpisodeRecordService
  def initialize(user:, episode:)
    @user = user
    @episode = episode
    @work = @episode.work
  end

  def call(episode_record_attributes, share_record: false)
    episode_record = user.episode_records.new(episode_record_attributes)
    episode_record.episode = episode
    episode_record.work = work
    episode_record.detect_locale!(:body)

    ActiveRecord::Base.transaction do
      episode_record.activity = user.create_or_last_activity!(episode_record, :create_episode_record)
      episode_record.record = user.records.create!(work: work)

      episode_record.save!

      user.update_share_record_status(share_record)
      episode.update_episode_record_bodies_count!(nil, episode_record)
      library_entry&.append_episode!(episode)

      if share_record
        ShareEpisodeRecordToTwitterJob.perform_later(user.id, episode_record.id)
      end
    end
  end

  private

  attr_reader :user, :episode, :work

  def library_entry
    @library_entry ||= user.library_entries.find_by(work: work)
  end
end
