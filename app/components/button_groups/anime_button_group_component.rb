# frozen_string_literal: true

module ButtonGroups
  class AnimeButtonGroupComponent < ApplicationV6Component
    def initialize(view_context, anime:, class_name: "")
      super view_context
      @anime = anime
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :div, class: "btn-group c-anime-button-group #{@class_name}" do
          h.html Dropdowns::AnimeStatusDropdownComponent.new(view_context, anime: @anime).render

          # TODO: アニメをお気に入りできるようにする
          # h.tag :button, type: "button", class: "btn btn-warning" do
          #   h.text "fav"
          # end

          h.tag :div, class: "btn-group c-anime-button-group__options" do
            h.tag :button, {
              class: "btn btn-outline-secondary dropdown-toggle",
              data_bs_toggle: "dropdown",
              type: "button"
            } do
              h.tag :i, class: "far fa-ellipsis-h"
            end

            h.tag :ul, class: "dropdown-menu" do
              h.tag :li do
                h.tag :button, {
                  class: "btn #{@class_name}",
                  type: "button"
                } do
                  h.text "menu item"
                end
              end
            end
          end
        end
      end
    end
  end
end
