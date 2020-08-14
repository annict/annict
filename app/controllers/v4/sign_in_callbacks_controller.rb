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

      session_interaction = SessionInteraction.find_by(kind: :sign_in, token: token)

      if !session_interaction || session_interaction.expired?
        @message = t("messages.sign_in_callback.show.expired_html").html_safe
        return
      end

      user = User.only_kept.find_by!(email: session_interaction.email)

      ActiveRecord::Base.transaction do
        unless user.confirmed_at?
          user.touch(:confirmed_at)
        end

        session_interaction.destroy

        sign_in user
      end

      flash[:notice] = t("messages.sign_in_callback.show.signed_in")
      redirect_to root_path
    end
  end
end
