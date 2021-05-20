# frozen_string_literal: true

module V6
  class SignUpController < V6::ApplicationController
    layout "v6/simple"

    before_action :redirect_if_signed_in

    def new
      @form = ::Forms::SignUpForm.new
      @recaptcha = Recaptcha.new(action: "sign_up")
    end
  end
end
