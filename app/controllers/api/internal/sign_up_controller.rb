# frozen_string_literal: true

module Api::Internal
  class SignUpController < Api::Internal::ApplicationController
    def create
      @form = Forms::SignUpForm.new(sign_up_form_params)
      @recaptcha = Recaptcha.new(action: "sign_up")

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

      unless @recaptcha.verify?(params[:recaptcha_token])
        return render json: [t("messages.recaptcha.not_verified")], status: :unprocessable_entity
      end

      EmailConfirmation.new(email: @form.email, back: @form.back).confirm_to_sign_up!

      render json: {flash: {type: :notice, message: t("messages.sign_up.create.mail_has_sent")}}, status: 201
    end

    private

    def sign_up_form_params
      permitted = params.require(:forms_sign_up_form).permit(:email)
      permitted[:back] = stored_location_for(:user)
      permitted
    end
  end
end
