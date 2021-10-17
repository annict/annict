# frozen_string_literal: true

module Updaters
  class WorkRecordUpdater
    attr_accessor :record

    def initialize(user:, form:)
      @user = user
      @form = form
      @record = @form.record
      @work_record = @record.work_record
    end

    def call
      @work_record.rating_overall_state = @form.rating_overall
      @work_record.rating_animation_state = @form.rating_animation
      @work_record.rating_music_state = @form.rating_music
      @work_record.rating_story_state = @form.rating_story
      @work_record.rating_character_state = @form.rating_character
      @work_record.modified_at = Time.zone.now
      @work_record.body = @form.comment
      @work_record.oauth_application = @form.oauth_application
      @work_record.detect_locale!(:body)
      @record.watched_at = @form.watched_at

      if @form.deprecated_title.present?
        @work_record.body = "#{@form.deprecated_title}\n\n#{@work_record.body}"
      end

      ActiveRecord::Base.transaction do
        @work_record.save!
        @record.save!
        @user.touch(:record_cache_expired_at)

        if @form.share_to_twitter
          @user.share_work_record_to_twitter(@work_record)
        end
      end

      self.record = @record

      self
    end
  end
end
