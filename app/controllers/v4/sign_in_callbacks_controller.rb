# frozen_string_literal: true

module V4
  class SignInCallbacksController < V4::ApplicationController
    layout "simple"

    def show
      redirect_if_signed_in

      token = params[:token]

      unless token
        return redirect_to root_path
      end

      confirmation = EmailConfirmation.find_by(event: :sign_in, token: token)

      if !confirmation || confirmation.expired?
        @message = t("messages.sign_in_callback.show.expired_html").html_safe
        return
      end

      user = User.only_kept.find_by!(email: confirmation.email)

      ActiveRecord::Base.transaction do
        unless user.confirmed_at?
          user.confirm
        end

        confirmation.destroy

        sign_in user
      end

      flash[:notice] = t("messages.sign_in_callback.show.signed_in")
      redirect_to(confirmation.back.presence || root_path)
    end
  end
end
