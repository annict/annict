class ConfirmationsController < ApplicationController
  def show
    user = User.confirm_by_token(params[:confirmation_token])

    if user.errors.empty?
      redirect_to root_path, notice: t('confirmations.confirmed')
    else
      redirect_to root_path, danger: t('confirmations.failure')
    end
  end
end
