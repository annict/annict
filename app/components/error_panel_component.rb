# frozen_string_literal: true

# TODO: form_with(local: false) にして ErrorPanelComponent2 に置き換える
class ErrorPanelComponent < ApplicationComponent
  def initialize(form:)
    @form = form
  end
end
