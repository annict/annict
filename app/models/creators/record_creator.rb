# frozen_string_literal: true

module Creators
  class RecordCreator
    attr_accessor :record

    def initialize(user:, form:)
      @user = user
      @form = form
      @episode = @form.episode
      @work = @episode&.work.presence || @form.work
    end

    def call
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
        library_entry.append_episode!(@episode)

        @user.update_share_record_setting(@form.share_to_twitter)
        @user.touch(:record_cache_expired_at)

        if @form.share_to_twitter
          @user.share_record_to_twitter(record)
        end
      end

      self.record = record

      self
    end
  end
end
