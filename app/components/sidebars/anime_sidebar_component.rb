# frozen_string_literal: true

module Sidebars
  class AnimeSidebarComponent < ApplicationComponent
    def initialize(anime_entity:, vod_channel_entities:)
      @anime_entity = anime_entity
      @vod_channel_entities = vod_channel_entities
    end

    private

    attr_reader :anime_entity, :vod_channel_entities
  end
end
