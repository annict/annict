# frozen_string_literal: true

module Settings
  class PasswordsController < ApplicationController
    before_action :authenticate_user!

    def update
      @user = User.find(current_user.id)

      @user.current_password = user_params[:current_password]
      return render_account_page unless @user.valid?(:password_check)

      @user.attributes = user_params.except(:current_password)
      return render_account_page unless @user.valid?(:password_update)

      @user.save(validate: false)
      bypass_sign_in(@user)

      redirect_to settings_account_path, notice: t("messages.accounts.updated")
    end

    private

    def render_account_page
      @user_email_form = Forms::UserEmailForm.new(email: current_user.email)

      render "/v3/settings/accounts/show"
    end

    def user_params
      params.require(:user).permit(:current_password, :password, :password_confirmation)
    end
  end
end
