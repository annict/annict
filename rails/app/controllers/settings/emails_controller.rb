# typed: false
# frozen_string_literal: true

module Settings
  class EmailsController < ApplicationV6Controller
    before_action :authenticate_user!

    def show
      @user = current_user
      @user_email_form = Forms::UserEmailForm.new(email: @user.email)
    end

    def update
      @user = current_user
      @user_email_form = Forms::UserEmailForm.new(user_email_form_params)

      if @user_email_form.invalid?
        return render :show, status: :unprocessable_entity
      end

      @user.confirm_to_update_email!(new_email: @user_email_form.email)

      flash[:notice] = t "messages.accounts.email_sent_for_confirmation"
      redirect_to settings_email_path
    end

    private

    def user_email_form_params
      params.require(:forms_user_email_form).permit(:email)
    end
  end
end
