# frozen_string_literal: true

module Updaters
  class EpisodeRecordUpdater
    attr_accessor :record

    def initialize(user:, form:)
      @user = user
      @form = form
    end

    def call
      @record = @form.record
      @episode_record = @record.episode_record

      @record.attributes = {
        oauth_application: @form.oauth_application,
        advanced_rating: @form.advanced_rating,
        body: @form.body,
        rating: @form.rating,
        watched_at: @form.watched_at.presence || @record.watched_at,
        modified_at: Time.zone.now
      }
      @record.detect_locale!(:body)

      ActiveRecord::Base.transaction do
        @record.save!

        @user.update_share_record_setting(@form.share_to_twitter)
        @user.touch(:record_cache_expired_at)

        if @form.share_to_twitter
          @user.share_record_to_twitter(record)
        end
      end

      self.record = @record

      self
    end
  end
end
