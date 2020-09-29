# frozen_string_literal: true

module ButtonGroups
  class RecordRatingButtonGroupComponent < ApplicationComponent
    def initialize(form:, input_name:)
      @form = form
      @input_name = input_name
    end
  end
end
