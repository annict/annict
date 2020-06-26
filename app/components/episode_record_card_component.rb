# frozen_string_literal: true

class EpisodeRecordCardComponent < ApplicationComponent
  def initialize(work_entity:, episode_entity:, episode_record_entity:)
    @work_entity = work_entity
    @episode_entity = episode_entity
    @episode_record_entity = episode_record_entity
  end

  private

  attr_reader :episode_entity, :episode_record_entity, :work_entity
end
