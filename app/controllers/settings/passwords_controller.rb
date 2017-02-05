# frozen_string_literal: true

module Settings
  class PasswordsController < ApplicationController
    permits :current_password, :password, :password_confirmation, model_name: "User"

    before_action :authenticate_user!

    def update(user)
      @user = User.find(current_user.id)

      @user.current_password = user[:current_password]
      return render("/accounts/show") unless @user.valid?(:password_check)

      @user.attributes = user.except(:current_password)
      return render("/accounts/show") unless @user.valid?(:password_update)

      @user.save(validate: false)
      bypass_sign_in(@user)

      redirect_to account_path, notice: t("messages.accounts.updated")
    end
  end
end
