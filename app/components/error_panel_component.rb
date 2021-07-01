# frozen_string_literal: true

class ErrorPanelComponent < ApplicationComponent
  def initialize(form:)
    @form = form
  end

  private

  attr_reader :form
end
