# frozen_string_literal: true

class TrackableService
  def initialize(user)
    @user = user
  end

  def latest_statuses
    LatestStatus.refresh_next_episode(@user)

    @user.
      latest_statuses.
      includes(:next_episode, :work).
      watching.
      has_next_episode.
      order(:position)
  end
end
