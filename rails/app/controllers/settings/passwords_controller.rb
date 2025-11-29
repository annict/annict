# typed: false
# frozen_string_literal: true

module Settings
  class PasswordsController < ApplicationV6Controller
    before_action :authenticate_user!

    def show
      @user = current_user
    end

    def update
      @user = User.find(current_user.id)

      @user.current_password = user_params[:current_password]
      if @user.invalid?(:password_check)
        return render :show, status: :unprocessable_entity
      end

      @user.attributes = user_params.except(:current_password)
      if @user.invalid?(:password_update)
        return render :show, status: :unprocessable_entity
      end

      @user.save(validate: false)
      bypass_sign_in(@user)

      flash[:notice] = t "messages._common.updated"
      redirect_to settings_password_path
    end

    private

    def user_params
      params.require(:user).permit(:current_password, :password, :password_confirmation)
    end
  end
end
