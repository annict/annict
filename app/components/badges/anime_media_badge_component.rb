# frozen_string_literal: true

module Badges
  class AnimeMediaBadgeComponent < ApplicationV6Component
    def initialize(view_context, anime:)
      super view_context
      @anime = anime
    end

    def render
      build_html do |h|
        h.tag :span, class: "badge rounded-pill u-bg-anime" do
          h.text @anime.media.text
        end
      end
    end
  end
end
