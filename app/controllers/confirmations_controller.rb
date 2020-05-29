# frozen_string_literal: true

class ConfirmationsController < ApplicationController
  def show
    user = User.confirm_by_token(params[:confirmation_token])

    if user.errors.empty?
      redirect_to new_user_session_path, notice: t("messages.confirmations.confirmed")
    else
      redirect_to root_path, alert: t("messages.confirmations.failure")
    end
  end
end
