# frozen_string_literal: true

module Updaters
  class AnimeRecordUpdater
    attr_accessor :record

    def initialize(user:, form:)
      @user = user
      @form = form
      @record = @form.record
      @anime_record = @record.anime_record
    end

    def call
      @anime_record.rating_overall_state = @form.rating_overall
      @anime_record.rating_animation_state = @form.rating_animation
      @anime_record.rating_music_state = @form.rating_music
      @anime_record.rating_story_state = @form.rating_story
      @anime_record.rating_character_state = @form.rating_character
      @anime_record.modified_at = Time.zone.now
      @anime_record.body = @form.comment
      @anime_record.oauth_application = @form.oauth_application
      @anime_record.detect_locale!(:body)

      if @form.deprecated_title.present?
        @anime_record.body = "#{@form.deprecated_title}\n\n#{@anime_record.body}"
      end

      ActiveRecord::Base.transaction do
        @anime_record.save!
        @record.touch
        @user.touch(:record_cache_expired_at)

        if @form.share_to_twitter
          @user.share_work_record_to_twitter(@anime_record)
        end
      end

      self.record = @record

      self
    end
  end
end
