# frozen_string_literal: true

class EpisodeRecordFooterComponent2 < ApplicationComponent
  def initialize(episode_record:)
    @episode_record = episode_record
    @record = @episode_record.record
  end
end
