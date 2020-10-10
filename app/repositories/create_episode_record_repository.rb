# frozen_string_literal: true

class CreateEpisodeRecordRepository < ApplicationRepository
  def execute(form:)
    result = mutate(variables: {
      episodeId: form.episode_id,
      comment: form.comment,
      rating: form.rating,
      shareToTwitter: form.share_to_twitter
    })

    if result.to_h["errors"]
      return [nil, MutationError.new(message: result.to_h["errors"][0]["message"])]
    end

    record_node = result.dig("data", "createEpisodeRecord", "record")

    [RecordEntity.from_node(record_node), nil]
  end
end
