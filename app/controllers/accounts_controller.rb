# frozen_string_literal: true

class AccountsController < ApplicationController
  before_action :authenticate_user!

  def show
    @user = current_user
    @user_email_form = UserEmailForm.new(email: current_user.email)
  end

  def update
    @user = User.find(current_user.id)
    @user.attributes = user_params

    if @user.save
      I18n.with_locale(@user.locale) do
        url = case I18n.locale.to_s
        when "ja" then ENV.fetch("ANNICT_URL")
        else
          ENV.fetch("ANNICT_EN_URL")
        end

        redirect_to("#{url}#{account_path}", notice: t("messages.accounts.updated"))
      end
    else
      render "/accounts/show"
    end
  end

  private

  def user_params
    params.require(:user).permit(:username, :time_zone, :locale, allowed_locales: [])
  end
end
