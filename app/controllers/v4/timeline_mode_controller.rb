# frozen_string_literal: true

module V4
  class TimelineModeController < V4::ApplicationController
    before_action :authenticate_user!

    def update
      current_user.setting.timeline_mode = permitted_params[:timeline_mode]

      if current_user.setting.save
        redirect_to root_path
      end
    end

    private

    def permitted_params
      params.permit(:timeline_mode)
    end
  end
end
