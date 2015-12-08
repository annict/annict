class AccountsController < ApplicationController
  permits :email

  before_action :authenticate_user!

  def update(user)
    current_user.email = user[:email]

    if current_user.valid?
      current_user.update_column(:unconfirmed_email, user[:email])
      current_user.resend_confirmation_instructions
      redirect_to account_path, notice: "確認メールを送信しました"
    else
      render "/accounts/show"
    end
  end
end
