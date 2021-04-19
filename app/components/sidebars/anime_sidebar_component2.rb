# frozen_string_literal: true

module Sidebars
  class AnimeSidebarComponent2 < ApplicationComponent
    def initialize(anime:, vod_channels:)
      super
      @anime = anime
      @vod_channels = vod_channels
    end
  end
end
