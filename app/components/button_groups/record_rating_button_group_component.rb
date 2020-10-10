# frozen_string_literal: true

module ButtonGroups
  class RecordRatingButtonGroupComponent < ApplicationComponent
    def initialize(form:, input_name:)
      @form = form
      @input_name = input_name
    end

    def button_class_name(rating)
      class_name = %w(btn)
      class_name << (@form.object.rating&.downcase == rating ? "u-btn-#{rating}" : "u-btn-outline-input-border")
      class_name.join(" ")
    end
  end
end
