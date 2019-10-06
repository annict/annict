# frozen_string_literal: true

class AccountsController < ApplicationController
  before_action :authenticate_user!

  def show
    @user = current_user
  end

  def update
    @user = User.find(current_user.id)
    @user.attributes = user_params

    if @user.valid?
      message = nil

      User.transaction do
        if @user.email_changed?
          @user.update_column(:unconfirmed_email, user_params[:email])
          @user.resend_confirmation_instructions
          message = t "messages.accounts.email_sent_for_confirmation"
        end

        @user.save(validate: false)
      end

      I18n.with_locale(@user.locale) do
        url = case I18n.locale.to_s
        when "ja" then ENV.fetch("ANNICT_JP_URL")
        else
          ENV.fetch("ANNICT_URL")
        end

        flash[:notice] = message.presence || t("messages.accounts.updated")
        redirect_to "#{url}#{account_path}"
      end
    else
      render "/accounts/show"
    end
  end

  private

  def user_params
    params.require(:user).permit(:username, :email, :time_zone, :locale, allowed_locales: [])
  end
end
