# frozen_string_literal: true

module Api
  module Internal
    class TrackedResourcesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def index
        render json: {
          episode_ids: current_user.episode_records.pluck(:episode_id),
          work_ids: current_user.work_records.pluck(:work_id)
        }
      end
    end
  end
end
