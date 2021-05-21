# frozen_string_literal: true

module V4
  class ErrorPanelComponent < V4::ApplicationComponent
    def initialize(form:)
      @form = form
    end

    private

    attr_reader :form
  end
end
