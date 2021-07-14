# frozen_string_literal: true

module Buttons
  class FavoriteButtonComponent < ApplicationV6Component
    def initialize(view_context, favoritable:, class_name: "")
      super view_context
      @favoritable = favoritable
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :button, {
          class: "btn c-favorite-button #{@class_name}",
          data_action: "favorite-button#toggle",
          data_controller: "favorite-button",
          data_favorite_button_default_class: "btn-outline-primary",
          data_favorite_button_favoriting_class: "btn-primary",
          data_favorite_button_favoritable_id_value: @favoritable.id,
          data_favorite_button_favoritable_type_value: @favoritable.class.name,
          type: "button"
        } do
          h.tag :i, class: "far fa-star"
        end
      end
    end
  end
end
