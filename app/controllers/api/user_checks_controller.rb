class Api::UserChecksController < Api::ApplicationController
  before_filter :authenticate_user!

  def index
    Check.refresh_episode(current_user)
    @checks = current_user.checks.has_episode
      .order(:position)
      .includes(:work, :episode)
  end

  def skip_episode(check_id)
    @check = Check.find(check_id)
    @check.skip_episode
  end
end
