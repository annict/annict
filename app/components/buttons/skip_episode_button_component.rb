# frozen_string_literal: true

module Buttons
  class SkipEpisodeButtonComponent < ApplicationV6Component
    def initialize(view_context, episode_id:, class_name: "")
      super view_context
      @episode_id = episode_id
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :button,
          class: "c-skip-episode-button #{skip_episode_button_class_name}",
          data_controller: "skip-episode-button",
          data_skip_episode_button_episode_id_value: @episode_id,
          data_skip_episode_button_loading_class: "c-skip-episode-button--loading",
          data_action: "click->skip-episode-button#skip" do
            h.tag :span, class: "c-skip-episode-button__spinner spinner-border spinner-border-sm"
            h.tag :i, class: "far fa-forward"
          end
      end
    end

    private

    def skip_episode_button_class_name
      classes = %w[btn]
      classes += @class_name.split(" ")
      classes.uniq.join(" ")
    end
  end
end
