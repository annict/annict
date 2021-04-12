# frozen_string_literal: true

module Buttons
  class SkipEpisodeButtonComponent < ApplicationComponent
    def initialize(episode_id:, class_name: "")
      @episode_id = episode_id
      @class_name = class_name
    end

    private

    def skip_episode_button_class_name
      classes = %w(btn)
      classes += @class_name.split(" ")
      classes.uniq.join(" ")
    end
  end
end
