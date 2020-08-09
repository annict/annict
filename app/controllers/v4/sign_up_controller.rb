# frozen_string_literal: true

module V4
  class SignUpController < V4::ApplicationController
    layout "simple"

    def new
      @form = SignUpForm.new
    end

    def create
      @form = SignUpForm.new(sign_up_form_params)

      return render(:new) unless @form.valid?

      flash[:notice] = t("messages.registrations.create.confirmation_mail_has_sent")
      redirect_to root_path
    end

    private

    def sign_up_form_params
      SignUpContract.new.call(params.to_unsafe_h["sign_up_form"])
    end
  end
end
