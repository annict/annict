# frozen_string_literal: true

module V4
  class SignUpController < V4::ApplicationController
    layout "simple"

    before_action :redirect_if_signed_in

    def new
      @form = SignUpForm.new
      @recaptcha = Recaptcha.new(action: "sign_up")
    end

    def create
      @form = SignUpForm.new(sign_up_form_attributes)
      @recaptcha = Recaptcha.new(action: "sign_up")

      unless @form.valid?
        return render(:new)
      end

      unless @recaptcha.verify?(params[:recaptcha_token])
        flash.now[:alert] = t("messages.recaptcha.not_verified")
        return render(:new)
      end

      EmailConfirmation.new(email: @form.email, back: @form.back).confirm_to_sign_up!

      flash[:notice] = t("messages.sign_up.create.mail_has_sent")
      redirect_to root_path
    end

    private

    def sign_up_form_attributes
      params.to_unsafe_h["sign_up_form"].merge(back: stored_location_for(:user))
    end
  end
end
