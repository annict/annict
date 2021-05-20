# frozen_string_literal: true

module V6
  class RegistrationsController < V6::ApplicationController
    layout "v6/simple"

    before_action :redirect_if_signed_in

    def new
      token = params[:token]

      unless token
        return redirect_to root_path
      end

      confirmation = EmailConfirmation.find_by(event: :sign_up, token: token)

      if !confirmation || confirmation.expired?
        @expired = true
        return
      end

      confirmation.touch(:expires_at)

      @form = ::Forms::RegistrationForm.new
      @form.email = confirmation.email
      @form.token = confirmation.token
    end
  end
end
