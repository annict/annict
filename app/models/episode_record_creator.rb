# frozen_string_literal: true

class EpisodeRecordCreator
  def initialize(user:, episode:, rating: nil, deprecated_rating: nil, comment: "", share_to_twitter: false) # rubocop:disable Metrics/ParameterLists
    @user = user
    @episode = episode
    @work = episode.work
    @rating = rating
    @deprecated_rating = deprecated_rating
    @comment = comment
    @share_to_twitter = share_to_twitter
  end

  def call
    result = Annict::Result.new(:record)
    episode_record = @episode.build_episode_record(
      user: @user,
      rating: @rating,
      deprecated_rating: @deprecated_rating,
      comment: @comment,
      share_to_twitter: @share_to_twitter
    )
    library_entry = @user.library_entries.find_by(work: @work)

    if episode_record.invalid?
      return result.failure(episode_record.errors.full_messages)
    end

    ActiveRecord::Base.transaction do
      episode_record.save!

      activity_group = @user.create_or_last_activity_group!(episode_record)
      @user.activities.create!(itemable: episode_record, activity_group: activity_group)

      @user.update_share_record_setting(@share_to_twitter)
      @user.touch(:record_cache_expired_at)
      library_entry&.append_episode!(@episode)

      if @user.share_record_to_twitter?
        @user.share_episode_record_to_twitter(episode_record)
      end
    end

    result.success(record: episode_record.record)
  end
end
