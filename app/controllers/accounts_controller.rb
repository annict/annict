# frozen_string_literal: true

class AccountsController < ApplicationController
  permits :username, :email, model_name: "User"

  before_action :authenticate_user!

  def show
    @user = current_user
    render layout: "v1/application"
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
          message = "確認メールを送信しました"
        end

        @user.save(validate: false)
      end

      redirect_to account_path, notice: (message.presence || "保存しました")
    else
      render "/accounts/show", layout: "v1/application"
    end
  end
end
