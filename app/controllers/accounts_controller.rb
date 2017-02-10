# frozen_string_literal: true

class AccountsController < ApplicationController
  permits :username, :email, :time_zone, :locale, model_name: "User"

  before_action :authenticate_user!

  def show
    @user = current_user
  end

  def update(user)
    @user = User.find(current_user.id)
    @user.attributes = user

    if @user.valid?
      message = nil

      User.transaction do
        if @user.email_changed?
          @user.update_column(:unconfirmed_email, user[:email])
          @user.resend_confirmation_instructions
          message = t "messages.accounts.email_sent_for_confirmation"
        end

        @user.save(validate: false)
      end

      I18n.locale = @user.locale
      redirect_to account_path, notice: (message.presence || t("messages.accounts.updated"))
    else
      render "/accounts/show"
    end
  end
end
