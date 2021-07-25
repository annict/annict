# frozen_string_literal: true

module Buttons
  class BulkWatchEpisodesButtonComponent < ApplicationV6Component
    def initialize(view_context, episode_id:, class_name: "")
      super view_context
      @episode_id = episode_id
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :button,
          class: "c-bulk-watch-episodes-button #{bulk_watch_episodes_button_class_name}",
          data_controller: "bulk-watch-episodes-button",
          data_bulk_watch_episodes_button_episode_id_value: @episode_id,
          data_bulk_watch_episodes_button_loading_class: "c-bulk-watch-episodes-button--loading",
          data_action: "click->bulk-watch-episodes-button#watch" do
            h.tag :span, class: "c-bulk-watch-episodes-button__spinner spinner-border spinner-border-sm"
            h.tag :i, class: "far fa-arrow-from-bottom"
          end
      end
    end

    private

    def bulk_watch_episodes_button_class_name
      classes = %w[btn]
      classes += @class_name.split(" ")
      classes.uniq.join(" ")
    end
  end
end
