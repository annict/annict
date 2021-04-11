# frozen_string_literal: true

module Buttons
  class WatchEpisodeButtonComponent < ApplicationComponent
    def initialize(episode_id:, page_category:, class_name: "")
      @episode_id = episode_id
      @page_category = page_category
      @class_name = class_name
    end

    private

    def watch_episode_button_class_name
      classes = %w(btn)
      classes += @class_name.split(" ")
      classes.uniq.join(" ")
    end
  end
end
