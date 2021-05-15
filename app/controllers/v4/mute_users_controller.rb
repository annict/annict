class MuteUsersController < ApplicationController
  before_action :authenticate_user!

  def destroy
    mute_user = current_user.mute_users.find(params[:id])
    mute_user.destroy
    redirect_to mutes_path, notice: "ミュートを解除しました"
  end
end
