# frozen_string_literal: true

class SignUpController < ApplicationV6Controller
  layout "main_simple"

  before_action :redirect_if_signed_in

  def new
    @form = SignUpForm.new
    @recaptcha = Recaptcha.new(action: "sign_up")
  end
end
