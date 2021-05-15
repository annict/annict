# frozen_string_literal: true

module V4::Sidebars
  class AnimeSidebarComponent < V4::ApplicationComponent
    def initialize(anime_entity:, vod_channel_entities:, viewer: nil)
      super
      @anime_entity = anime_entity
      @vod_channel_entities = vod_channel_entities
      @viewer = viewer
    end
  end
end
