# frozen_string_literal: true

class RecaptchaInputComponent < ApplicationComponent
  def initialize(recaptcha:)
    @recaptcha = recaptcha
  end
end
