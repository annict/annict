# frozen_string_literal: true

module V4
  class UserEmailsController < V4::ApplicationController
    before_action :authenticate_user!

    def update
      @user_email_form = UserEmailForm.new(user_email_form_params)

      unless @user_email_form.valid?
        @user = current_user
        return render("/accounts/show")
      end

      current_user.confirm_to_update_email!(new_email: @user_email_form.email)

      redirect_to(root_path, notice: t("messages.accounts.email_sent_for_confirmation"))
    end

    private

    def user_email_form_params
      UserEmailContract.new.call(params.to_unsafe_h["user_email_form"])
    end
  end
end
