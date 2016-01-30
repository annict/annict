module Api
  class LatestStatusesController < Api::ApplicationController
    before_action :authenticate_user!

    def index
      LatestStatus.refresh_next_episode(current_user)
      @latest_statuses = current_user.
        latest_statuses.
        has_next_episode.
        order(:position)
    end

    def skip_episode(latest_status_id)
      @latest_status = LatestStatus.find(latest_status_id)
      @latest_status.append_episode(@latest_status.next_episode)
    end
  end
end
