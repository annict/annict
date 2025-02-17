# typed: false
# frozen_string_literal: true

module Deprecated::Buttons
  class StarButtonComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, starrable:, class_name: "")
      super view_context
      @starrable = starrable
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :button, {
          class: "btn c-star-button #{@class_name}",
          data_action: "star-button#toggle",
          data_controller: "star-button",
          data_star_button_default_class: "btn-outline-warning",
          data_star_button_starred_class: "btn-warning",
          data_star_button_starrable_id_value: @starrable.id,
          data_star_button_starrable_type_value: @starrable.class.name,
          type: "button"
        } do
          h.tag :i, class: "fa-solid fa-star"
        end
      end
    end
  end
end
