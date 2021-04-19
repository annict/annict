# frozen_string_literal: true

class EpisodeRecordCardComponent2 < ApplicationComponent
  def initialize(episode_record:)
    @episode_record = episode_record
    @anime = @episode_record.anime
    @episode = @episode_record.episode
  end
end
