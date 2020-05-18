# frozen_string_literal: true

class CreateEpisodeRecordRepository < ApplicationRepository
  def create(episode:, params:)
    result = execute(variables: {
      episodeId: Canary::AnnictSchema.id_from_object(episode, Episode),
      body: params[:body],
      rating: params[:rating]&.to_f,
      ratingState: params[:rating_state]&.upcase.presence || nil,
      shareToTwitter: params[:share_to_twitter].in?(%w(true 1))
    })

    if result.to_h["errors"]
      return [nil, MutationError.new(message: result.to_h["errors"][0]["message"])]
    end

    data = result.dig("data", "createEpisodeRecord", "episodeRecord")
    entity = EpisodeRecordEntity.new(
      id: data["annictId"]
    )

    [entity, nil]
  end
end
