# typed: false
# frozen_string_literal: true

module Api::Internal
  class SignInController < Api::Internal::ApplicationController
    def create
      @form = Forms::SignInForm.new(sign_in_form_params)
      @recaptcha = Recaptcha.new(action: "sign_in")

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

      unless @recaptcha.verify?(params[:recaptcha_token])
        return render json: [t("messages.recaptcha.not_verified")], status: :unprocessable_entity
      end

      EmailConfirmation.new(email: @form.email, back: @form.back).confirm_to_sign_in!

      render json: {flash: {type: :notice, message: t("messages.sign_in.create.mail_has_sent")}}, status: 201
    end

    private

    def sign_in_form_params
      params.require(:forms_sign_in_form).permit(:email).merge(back: stored_location_for(:user))
    end
  end
end
