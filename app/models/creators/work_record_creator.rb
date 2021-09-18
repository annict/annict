# frozen_string_literal: true

module Creators
  class WorkRecordCreator
    attr_accessor :record

    def initialize(user:, form:)
      @user = user
      @form = form
      @work = @form.work
    end

    def call
      record = @user.records.new(
        work: @work,
        oauth_application: @form.oauth_application,
        body: @form.body,
        rating: @form.rating,
        animation_rating: @form.animation_rating,
        character_rating: @form.character_rating,
        music_rating: @form.music_rating,
        story_rating: @form.story_rating,
        watched_at: @form.watched_at.presence || Time.zone.now
      )
      record.recordable = WorkRecord.new
      record.detect_locale!(:body)

      ActiveRecord::Base.transaction do
        record.save!

        if @form.watched_at.nil?
          activity_group = @user.create_or_last_activity_group!(record)
          @user.activities.create!(itemable: record, activity_group: activity_group)
        end

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
