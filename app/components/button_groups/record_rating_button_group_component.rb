# frozen_string_literal: true

module ButtonGroups
  class RecordRatingButtonGroupComponent < ApplicationComponent
    def initialize(form:, rating_kind:)
      @form = form
      @rating_kind = rating_kind
    end

    def input_name
      "#{@form.object.class.name.underscore}[#{@rating_kind}]"
    end

    def rating
      @rating ||= @form.object.send(@rating_kind)&.downcase
    end

    def button_class_name(rating_state)
      class_name = %w(btn)
      class_name << (rating&.downcase == rating_state ? "u-btn-#{rating_state}" : "u-btn-outline-input-border")
      class_name.join(" ")
    end
  end
end
