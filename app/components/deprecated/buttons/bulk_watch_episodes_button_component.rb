# typed: false
# frozen_string_literal: true

module Deprecated::Buttons
  class BulkWatchEpisodesButtonComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, episode_id:, button_text: "", class_name: "")
      super view_context
      @episode_id = episode_id
      @button_text = button_text
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
            h.tag :i, class: "fa-solid fa-arrow-from-bottom"

            if @button_text.present?
              h.tag :span, class: "ms-1" do
                h.text @button_text
              end
            end
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
