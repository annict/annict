# frozen_string_literal: true

module ButtonGroups
  class AnimeButtonGroupComponent < ApplicationV6Component
    def initialize(view_context, anime:, class_name: "", show_option_button: true)
      super view_context
      @anime = anime
      @class_name = class_name
      @show_option_button = show_option_button
    end

    def render
      build_html do |h|
        h.tag :div, class: "btn-group c-anime-button-group #{@class_name}" do
          h.html Dropdowns::StatusSelectDropdownComponent.new(view_context, anime: @anime).render

          # TODO: アニメをお気に入りできるようにする
          # h.tag :button, type: "button", class: "btn btn-outline-warning" do
          #   h.tag :i, class: "far fa-star"
          # end

          if @show_option_button
            h.tag :button, {
              class: "btn btn-outline-secondary",
              data_controller: "tracking-offcanvas-button",
              data_tracking_offcanvas_button_anime_id_value: @anime.id,
              data_tracking_offcanvas_button_frame_path: view_context.fragment_trackable_anime_path(@anime.id),
              data_action: "click->tracking-offcanvas-button#open",
              type: "button"
            } do
              h.tag :i, class: "far fa-ellipsis-h"
            end
          end
        end
      end
    end
  end
end
