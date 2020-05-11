# frozen_string_literal: true

class CreateEpisodeRecordService
  def initialize(user:, episode:)
    @user = user
    @episode = episode
    @work = @episode.work
  end

  def call(episode_record_attributes, share_record_to_twitter: false)
    episode_record = user.episode_records.new(episode_record_attributes)
    episode_record.episode = episode
    episode_record.work = work
    episode_record.detect_locale!(:body)

    ActiveRecord::Base.transaction do
      episode_record.record = user.records.create!(work: work)
      episode_record.activity = user.build_or_last_activity(episode_record, episode, :create_episode_record)
      persisted_activity = episode_record.activity.persisted?

      episode_record.save!

      if persisted_activity
        user.create_repetitive_activity!(episode_record, episode, :create_episode_record)
      end

      user.update_share_record_setting(share_record_to_twitter)
      episode.update_record_body_count!(nil, episode_record, field: :episode_record_bodies_count)
      library_entry&.append_episode!(episode)

      if user.share_record_to_twitter?
        user.share_episode_record_to_twitter(episode_record)
      end
    end
  end

  private

  attr_reader :user, :episode, :work

  def library_entry
    @library_entry ||= user.library_entries.find_by(work: work)
  end
end
