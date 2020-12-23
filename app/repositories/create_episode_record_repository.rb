# frozen_string_literal: true

class CreateEpisodeRecordRepository < ApplicationRepository
  class RepositoryResult < Result
    attr_accessor :record_entity
  end

  def execute(form:)
    data = mutate(
      variables: {
        episodeId: form.episode_id,
        comment: form.comment,
        rating: form.rating,
        shareToTwitter: form.share_to_twitter
      }
    )
    result = validate(data)

    if result.success?
      record_node = data.dig("data", "createEpisodeRecord", "record")
      result.record_entity = RecordEntity.from_node(record_node)
    end

    result
  end

  private

  def result_class
    RepositoryResult
  end
end
