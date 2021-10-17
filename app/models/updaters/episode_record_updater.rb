# frozen_string_literal: true

module Updaters
  class EpisodeRecordUpdater
    attr_accessor :record

    def initialize(user:, form:)
      @user = user
      @form = form
      @record = @form.record
      @episode_record = @record.episode_record
    end

    def call
      @episode_record.rating_state = @form.rating&.downcase
      @episode_record.modify_body = @episode_record.body != @form.comment
      @episode_record.body = @form.comment
      @episode_record.oauth_application = @form.oauth_application
      @episode_record.detect_locale!(:body)
      @record.watched_at = @form.watched_at

      ActiveRecord::Base.transaction do
        @episode_record.save!
        @record.save!
        @user.touch(:record_cache_expired_at)

        if @form.share_to_twitter
          @user.share_episode_record_to_twitter(@episode_record)
        end
      end

      self.record = @form.record

      self
    end
  end
end
