# frozen_string_literal: true

module V4
  class SignUpController < V4::ApplicationController
    layout "simple"

    def new
      redirect_if_signed_in

      @form = SignUpForm.new
    end

    def create
      redirect_if_signed_in

      @form = SignUpForm.new(sign_up_form_params)

      return render(:new) unless @form.valid?

      SessionInteraction.start_sign_up!(email: @form.email, locale: I18n.locale)

      flash[:notice] = t("messages.sign_up.create.mail_has_sent")
      redirect_to root_path
    end

    private

    def sign_up_form_params
      SignUpContract.new.call(params.to_unsafe_h["sign_up_form"])
    end
  end
end
