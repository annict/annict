# frozen_string_literal: true

module Api
  module Internal
    class LatestStatusesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def show
        @latest_status = current_user.latest_statuses.find_by(work_id: params[:work_id])
        @user = current_user
      end

      def skip_episode
        @latest_status = LatestStatus.find(params[:latest_status_id])
        @latest_status.append_episode!(@latest_status.next_episode)
        @user = current_user
        render :show
      end
    end
  end
end
