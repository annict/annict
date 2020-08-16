# frozen_string_literal: true

class CreateAnimeRecordRepository < ApplicationRepository
  def create(anime:, params:)
    result = execute(variables: {
      animeId: Canary::AnnictSchema.id_from_object(anime, Work),
      body: params[:body],
      ratingOverallState: params[:rating_overall_state]&.upcase.presence || nil,
      ratingAnimationState: params[:rating_animation_state]&.upcase.presence || nil,
      ratingMusicState: params[:rating_music_state]&.upcase.presence || nil,
      ratingStoryState: params[:rating_story_state]&.upcase.presence || nil,
      ratingCharacterState: params[:rating_character_state]&.upcase.presence || nil,
      shareToTwitter: params[:share_to_twitter].in?(%w(true 1))
    })

    if result.to_h["errors"]
      return [nil, MutationError.new(message: result.to_h["errors"][0]["message"])]
    end

    anime_record_node = result.dig("data", "createAnimeRecord", "animeRecord")

    [AnimeRecordEntity.from_node(anime_record_node), nil]
  end
end
