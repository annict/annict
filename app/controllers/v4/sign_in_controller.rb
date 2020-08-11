# frozen_string_literal: true

module V4
  class SignInController < V4::ApplicationController
    layout "simple"

    def new
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
      @form = SignInForm.new(sign_in_form_params)

      return render(:new) unless @form.valid?

      SessionInteraction.start_sign_in!(email: @form.email, locale: I18n.locale)

      flash[:notice] = t("messages.sign_in.create.mail_has_sent")
      redirect_to root_path
    end

    private

    def sign_in_form_params
      SignInContract.new.call(params.to_unsafe_h["sign_in_form"])
    end
  end
end
