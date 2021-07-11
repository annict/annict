# frozen_string_literal: true

module Settings
  class EmailCallbacksController < ApplicationV6Controller
    before_action :authenticate_user!

    def show
      token = params[:token]

      unless token
        return redirect_to(root_path)
      end

      confirmation = current_user.email_confirmations.find_by(event: :update_email, token: token)

      if !confirmation || confirmation.expired?
        return redirect_to(root_path, alert: t("messages.user_email_callbacks.show.expired"))
      end

      ActiveRecord::Base.transaction do
        current_user.update!(email: confirmation.email)
        confirmation.destroy
      end

      redirect_to(root_path, notice: t("messages.user_email_callbacks.show.updated"))
    end
  end
end
