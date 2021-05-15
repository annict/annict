# frozen_string_literal: true

class CreateAnimeRecordService < ApplicationService
  class CreateAnimeRecordServiceResult < Result
    attr_accessor :record
  end

  def initialize( # rubocop:disable Metrics/ParameterLists
    user:,
    anime:,
    rating_overall: nil,
    rating_animation: nil,
    rating_music: nil,
    rating_story: nil,
    rating_character: nil,
    comment: "",
    share_to_twitter: false
  )
    super()
    @user = user
    @anime = anime
    @rating_overall = rating_overall
    @rating_animation = rating_animation
    @rating_music = rating_music
    @rating_story = rating_story
    @rating_character = rating_character
    @comment = comment
    @share_to_twitter = share_to_twitter
  end

  def call
    anime_record = @anime.build_anime_record(
      user: @user,
      rating_overall: @rating_overall,
      rating_animation: @rating_animation,
      rating_music: @rating_music,
      rating_story: @rating_story,
      rating_character: @rating_character,
      comment: @comment,
      share_to_twitter: @share_to_twitter
    )

    ActiveRecord::Base.transaction do
      unless anime_record.save
        @result.errors.concat(anime_record.errors.full_messages.map { |msg| Result::Error.new(message: msg) })
        raise ActiveRecord::Rollback
      end

      activity_group = @user.create_or_last_activity_group!(anime_record)
      @user.activities.create!(itemable: anime_record, activity_group: activity_group)

      @user.update_share_record_setting(@share_to_twitter)
      @user.touch(:record_cache_expired_at)

      if @user.share_record_to_twitter?
        @user.share_work_record_to_twitter(anime_record)
      end
    end

    if @result.success?
      @result.record = anime_record.record
    end

    @result
  end

  private

  def result_class
    CreateAnimeRecordServiceResult
  end
end
