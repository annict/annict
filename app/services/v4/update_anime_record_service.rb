# frozen_string_literal: true

module V4
  class UpdateAnimeRecordService < V4::ApplicationService
    class ServiceResult < Result
      attr_accessor :record
    end

    def initialize(
      user:,
      record:,
      rating_overall: nil,
      rating_animation: nil,
      rating_music: nil,
      rating_story: nil,
      rating_character: nil,
      comment: "",
      share_to_twitter: false,
      oauth_application: nil
    ) # rubocop:disable Metrics/ParameterLists
      super()
      @user = user
      @record = record
      @anime_record = @record.anime_record
      @rating_overall = rating_overall
      @rating_animation = rating_animation
      @rating_music = rating_music
      @rating_story = rating_story
      @rating_character = rating_character
      @comment = comment
      @share_to_twitter = share_to_twitter
      @oauth_application = oauth_application
    end

    def call
      @anime_record.rating_overall_state = @rating_overall&.downcase
      @anime_record.rating_animation_state = @rating_animation&.downcase
      @anime_record.rating_music_state = @rating_music&.downcase
      @anime_record.rating_story_state = @rating_story&.downcase
      @anime_record.rating_character_state = @rating_character&.downcase
      @anime_record.modified_at = Time.zone.now
      @anime_record.body = @comment
      @anime_record.oauth_application = @oauth_application
      @anime_record.detect_locale!(:body)

      ActiveRecord::Base.transaction do
        @anime_record.save!
        @user.touch(:record_cache_expired_at)

        if @share_to_twitter
          @user.share_work_record_to_twitter(@anime_record)
        end
      end

      if @result.success?
        @result.record = @record
      end

      @result
    end

    private

    def result_class
      ServiceResult
    end
  end
end
