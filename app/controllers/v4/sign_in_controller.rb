# frozen_string_literal: true

module V4
  class SignInController < V4::ApplicationController
    layout "simple"

    def new
      redirect_if_signed_in

      if params[:back]
        store_location_for(:user, params[:back])
      end

      # From OAuth client
      if params[:client_id]
        @oauth_app = Doorkeeper::Application.available.find_by(uid: params[:client_id])
      end

      @form = SignInForm.new
    end

    def create
      redirect_if_signed_in

      @form = SignInForm.new(sign_in_form_params)

      return render(:new) unless @form.valid?

      EmailConfirmation.new(email: @form.email, back: @form.back).confirm_to_sign_in!

      flash[:notice] = t("messages.sign_in.create.mail_has_sent")
      redirect_to root_path
    end

    private

    def sign_in_form_params
      attributes = params.to_unsafe_h["sign_in_form"].merge(back: stored_location_for(:user))

      SignInContract.new.call(attributes)
    end
  end
end
