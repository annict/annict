# frozen_string_literal: true

module ButtonGroups
  class RecordRatingButtonGroupComponent < ApplicationComponent
    def initialize(form:, rating_field:)
      @form = form
      @rating_field = rating_field
    end

    def input_name
      "#{@form.object.class.name.underscore}[#{@rating_field}]"
    end

    def rating
      @rating ||= @form.object.send(@rating_field)&.downcase.presence
    end

    def button_class_name(rating_state)
      class_name = %w[btn]
      class_name << (rating&.downcase == rating_state ? "u-btn-#{rating_state}" : "u-btn-outline-input-border")
      class_name.join(" ")
    end
  end
end
