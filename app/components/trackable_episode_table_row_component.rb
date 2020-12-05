# frozen_string_literal: true

class TrackableEpisodeTableRowComponent < ApplicationComponent
  def initialize(episode_entity:)
    @episode_entity = episode_entity
  end
end
