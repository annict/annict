# frozen_string_literal: true

module Creators
  class RecordCreator
    attr_accessor :record

    def initialize(user:, form:)
      @user = user
      @form = form
    end

    def call
      @episode = @form.episode
      @work = @episode&.work.presence || @form.work
      @share_to_twitter = @form.instant ? @user.share_record_to_twitter? : @form.share_to_twitter

      record = @user.records.new(
        work: @work,
        episode: @episode,
        rating: @form.rating,
        advanced_rating: @form.advanced_rating,
        body: @form.body,
        watched_at: Time.zone.now
      )
      record.detect_locale!(:body)

      ActiveRecord::Base.transaction do
        record.save!

        activity_group = @user.create_or_last_activity_group!(record)
        @user.activities.create!(itemable: record, activity_group: activity_group)

        library_entry = @user.library_entries.where(work: @work).first_or_create!

        if @episode
          library_entry.append_episode!(@episode)
        end

        @user.update_share_record_setting(@share_to_twitter)
        @user.touch(:record_cache_expired_at)

        if @share_to_twitter
          @user.share_record_to_twitter(record)
        end
      end

      self.record = record

      self
    end
  end
end
