# frozen_string_literal: true

module Settings
  class AccountsController < ApplicationV6Controller
    before_action :authenticate_user!

    def show
      @user = current_user
      @user_email_form = Forms::UserEmailForm.new(email: current_user.email)
    end

    def update
      @user = User.find(current_user.id)
      @user.attributes = user_params

      if @user.save
        I18n.with_locale(@user.locale) do
          url = case I18n.locale.to_s
          when "ja" then ENV.fetch("ANNICT_JP_URL")
          else
            ENV.fetch("ANNICT_URL")
          end

          redirect_to("#{url}#{settings_account_path}", notice: t("messages.accounts.updated"))
        end
      else
        render "/v3/settings/accounts/show"
      end
    end

    private

    def user_params
      params.require(:user).permit(:username, :time_zone, :locale, allowed_locales: [])
    end
  end
end
