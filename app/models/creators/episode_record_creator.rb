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
      record = @user.records.new(
        work: @work,
        episode: @episode,
        oauth_application: @form.oauth_application,
        advanced_rating: @form.advanced_rating,
        body: @form.body,
        rating: @form.rating,
        watched_at: @form.watched_at.presence || Time.zone.now
      )
      record.recordable = EpisodeRecord.new
      record.detect_locale!(:body)

      ActiveRecord::Base.transaction do
        record.save!

        activity_group = @user.create_or_last_activity_group!(record)
        @user.activities.create!(itemable: record, activity_group: activity_group)

        library_entry = @user.library_entries.where(work: @work).first_or_create!
        library_entry.append_episode!(@episode)

        @user.update_share_record_setting(@form.share_to_twitter)
        @user.touch(:record_cache_expired_at)

        if @user.share_record_to_twitter?
          @user.share_record_to_twitter(record)
        end
      end

      self.record = record

      self
    end
  end
end
