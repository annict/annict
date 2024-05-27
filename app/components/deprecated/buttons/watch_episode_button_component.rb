# typed: false
# frozen_string_literal: true

module Deprecated::Buttons
  class WatchEpisodeButtonComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, episode_id:, button_text: "", class_name: "")
      super view_context
      @episode_id = episode_id
      @button_text = button_text
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :button, {
          class: "c-watch-episode-button #{watch_episode_button_class_name}",
          data_controller: "watch-episode-button",
          data_watch_episode_button_episode_id_value: @episode_id,
          data_watch_episode_button_loading_class: "c-watch-episode-button--loading",
          data_watch_episode_button_page_category_value: page_category,
          data_action: "click->watch-episode-button#watch"
        } do
          h.tag :span, class: "c-watch-episode-button__spinner spinner-border spinner-border-sm"
          h.tag :i, class: "far fa-check"

          if @button_text.present?
            h.tag :span, class: "ms-1" do
              h.text @button_text
            end
          end
        end
      end
    end

    private

    def watch_episode_button_class_name
      classes = %w[btn]
      classes += @class_name.split(" ")
      classes.uniq.join(" ")
    end
  end
end
