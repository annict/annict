# frozen_string_literal: true

class CreateAnimeRecordRepository < ApplicationRepository
  def execute(form:)
    result = mutate(variables: {
      animeId: form.anime_id,
      comment: form.comment,
      ratingOverall: form.rating_overall&.upcase.presence || nil,
      ratingAnimation: form.rating_animation&.upcase.presence || nil,
      ratingMusic: form.rating_music&.upcase.presence || nil,
      ratingStory: form.rating_story&.upcase.presence || nil,
      ratingCharacter: form.rating_character&.upcase.presence || nil,
      shareToTwitter: form.share_to_twitter
    })

    if result.to_h["errors"]
      return [nil, MutationError.new(message: result.to_h["errors"][0]["message"])]
    end

    record_node = result.dig("data", "createAnimeRecord", "record")

    [RecordEntity.from_node(record_node), nil]
  end
end
