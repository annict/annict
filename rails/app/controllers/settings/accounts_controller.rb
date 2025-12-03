# typed: false
# frozen_string_literal: true

module Settings
  class AccountsController < ApplicationV6Controller
    before_action :authenticate_user!

    def show
      @user = current_user
    end

    def update
      @user = User.find(current_user.id)
      @user.attributes = user_params

      if @user.invalid?
        return render :show, status: :unprocessable_entity
      end

      @user.save!

      url = case @user.locale.to_s
      when "ja"
        ENV.fetch("ANNICT_URL")
      else
        ENV.fetch("ANNICT_EN_URL")
      end

      flash[:notice] = t "messages._common.updated"
      redirect_to "#{url}#{settings_account_path}", allow_other_host: true
    end

    private

    def user_params
      params.require(:user).permit(:username, :time_zone, :locale, allowed_locales: [])
    end
  end
end
