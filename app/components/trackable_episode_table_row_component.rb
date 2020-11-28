# frozen_string_literal: true

class TrackableEpisodeTableRowComponent < ApplicationComponent
  def initialize(episode_entity:)
    @episode_entity = episode_entity
  end

  def class_name
    classes = []

    if @episode_entity.viewer_did_track_in_current_status
      classes << "c-trackable-episode-table-row--tracked"
    end

    classes.join(" ")
  end
end