# frozen_string_literal: true

module Deprecated::ButtonGroups
  class RecordRatingButtonGroupComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, form:, rating_field:)
      super view_context
      @form = form
      @rating_field = rating_field
    end

    def render
      build_html do |h|
        h.tag :div, {
          class: "c-record-rating-button-group",
          data_controller: "record-rating"
        } do
          h.tag :div, class: "btn-group btn-group-sm" do
            EpisodeRecord.rating_state.values.each do |rating_state|
              h.tag :div, {
                class: button_class_name(rating_state),
                data_action: "click->record-rating#changeState",
                data_state: rating_state
              } do
                h.html view_context.rating_state_icon(rating_state, class: "me-1")
                h.text rating_state.text
              end
            end
          end

          h.html view_context.hidden_field_tag(
            input_name,
            rating,
            "data-record-rating-target": "input"
          )
        end
      end
    end

    private

    def input_name
      "#{@form.object.class.name.underscore.tr("/", "_")}[#{@rating_field}]"
    end

    def rating
      @rating ||= @form.object.send(@rating_field)&.downcase.presence
    end

    def button_class_name(rating_state)
      class_name = %w[btn text-body]
      class_name << (rating&.downcase == rating_state ? "u-btn-#{rating_state}" : "btn-outline-secondary")
      class_name.join(" ")
    end
  end
end
