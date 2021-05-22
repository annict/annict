# frozen_string_literal: true

module V6::Settings
  class MutedUsersController < V6::ApplicationController
    before_action :authenticate_user!

    def destroy
      mute_user = current_user.mute_users.find(params[:mute_user_id])
      mute_user.destroy
      redirect_to settings_muted_user_list_path, notice: "ミュートを解除しました"
    end
  end
end
