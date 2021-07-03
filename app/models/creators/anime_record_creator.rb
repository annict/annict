# frozen_string_literal: true

module Creators
  class AnimeRecordCreator
    attr_accessor :record

    def initialize(user:, form:)
      @user = user
      @form = form
      @anime = @form.anime
    end

    def call
      anime_record = @anime.build_anime_record(
        user: @user,
        rating_overall: @form.rating_overall,
        rating_animation: @form.rating_animation,
        rating_music: @form.rating_music,
        rating_story: @form.rating_story,
        rating_character: @form.rating_character,
        comment: @form.comment,
        share_to_twitter: @form.share_to_twitter
      )

      if @form.deprecated_title.present?
        anime_record.body = "#{@form.deprecated_title}\n\n#{anime_record.body}"
      end

      ActiveRecord::Base.transaction do
        anime_record.save!

        activity_group = @user.create_or_last_activity_group!(anime_record)
        @user.activities.create!(itemable: anime_record, activity_group: activity_group)

        @user.update_share_record_setting(@form.share_to_twitter)
        @user.touch(:record_cache_expired_at)

        if @user.share_record_to_twitter?
          @user.share_work_record_to_twitter(anime_record)
        end
      end

      self.record = anime_record.record

      self
    end
  end
end
