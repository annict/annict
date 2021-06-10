# frozen_string_literal: true

class AnimeSidebarComponent < ApplicationComponent
  def initialize(anime_entity:, vod_channel_entities:, viewer: nil)
    super
    @anime_entity = anime_entity
    @vod_channel_entities = vod_channel_entities
    @viewer = viewer
  end
end
