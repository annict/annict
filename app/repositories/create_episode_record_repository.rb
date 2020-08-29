# frozen_string_literal: true

class CreateEpisodeRecordRepository < ApplicationRepository
  def execute(episode:, params:)
    result = mutate(variables: {
      episodeId: Canary::AnnictSchema.id_from_object(episode, Episode),
      body: params[:body],
      rating: params[:rating]&.to_f,
      ratingState: params[:rating_state]&.upcase.presence || nil,
      shareToTwitter: params[:share_to_twitter].in?(%w(true 1))
    })

    if result.to_h["errors"]
      return [nil, MutationError.new(message: result.to_h["errors"][0]["message"])]
    end

    episode_record_node = result.dig("data", "createEpisodeRecord", "episodeRecord")

    [EpisodeRecordEntity.from_node(episode_record_node), nil]
  end
end
