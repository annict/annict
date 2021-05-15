# frozen_string_literal: true

module V6
  class RecaptchaInputComponent < V6::ApplicationComponent
    def initialize(view_context, recaptcha:)
      super view_context
      @recaptcha = recaptcha
    end

    def render
      build_html do |h|
        h.tag :input, name: "recaptcha_token", type: "hidden", id: @recaptcha.id
      end
    end
  end
end
