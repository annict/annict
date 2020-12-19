# frozen_string_literal: true

class CreateEpisodeRecordService < ApplicationService
  class CreateEpisodeRecordServiceResult < Result
    attr_accessor :record
  end

  def initialize(user:, episode:, rating: nil, comment: "", share_to_twitter: false)
    super()
    @user = user
    @episode = episode
    @work = episode.work
    @rating = rating
    @comment = comment
    @share_to_twitter = share_to_twitter
  end

  def call
    episode_record = @episode.build_episode_record(
      user: @user,
      rating: @rating,
      comment: @comment,
      share_to_twitter: @share_to_twitter
    )
    library_entry = @user.library_entries.find_by(work: @work)

    ActiveRecord::Base.transaction do
      unless episode_record.save
        @result.errors.concat(episode_record.errors.full_messages.map { |msg| Result::Error.new(message: msg) })
        raise ActiveRecord::Rollback
      end

      activity_group = @user.create_or_last_activity_group!(episode_record)
      @user.activities.create!(itemable: episode_record, activity_group: activity_group)

      @user.update_share_record_setting(@share_to_twitter)
      library_entry&.append_episode!(@episode)

      if @user.share_record_to_twitter?
        @user.share_episode_record_to_twitter(episode_record)
      end
    end

    if @result.success?
      @result.record = episode_record.record
    end

    @result
  end

  private

  def result_class
    CreateEpisodeRecordServiceResult
  end
end
