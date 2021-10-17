# frozen_string_literal: true

module Creators
  class EpisodeRecordCreator
    attr_accessor :record

    def initialize(user:, form:)
      @user = user
      @form = form
      @episode = @form.episode
      @work = @episode.work
    end

    def call
      episode_record = @episode.build_episode_record(
        user: @user,
        rating: @form.rating,
        deprecated_rating: @form.deprecated_rating,
        comment: @form.comment,
        share_to_twitter: @form.share_to_twitter,
        watched_at: @form.watched_at
      )

      ActiveRecord::Base.transaction do
        episode_record.save!

        if @form.create_activity?
          activity_group = @user.create_or_last_activity_group!(episode_record)
          @user.activities.create!(itemable: episode_record, activity_group: activity_group)
        end

        library_entry = @user.library_entries.where(work: @work).first_or_create!
        library_entry.append_episode!(@episode)

        @user.update_share_record_setting(@form.share_to_twitter)
        @user.touch(:record_cache_expired_at)

        if @user.share_record_to_twitter?
          @user.share_episode_record_to_twitter(episode_record)
        end
      end

      self.record = episode_record.record

      self
    end
  end
end
