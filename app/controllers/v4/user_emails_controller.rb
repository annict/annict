# frozen_string_literal: true

module V4
  class UserEmailsController < ApplicationController
    before_action :authenticate_user!

    def update
      @user_email_form = UserEmailForm.new(user_email_form_params)

      if @user_email_form.invalid?
        @user = current_user
        return render("/accounts/show")
      end

      current_user.confirm_to_update_email!(new_email: @user_email_form.email)

      redirect_to(root_path, notice: t("messages.accounts.email_sent_for_confirmation"))
    end

    private

    def user_email_form_params
      params.require(:user_email_form).permit(:email)
    end
  end
end
