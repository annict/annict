class UserEpisodesQuery
  def initialize(user)
    @user = user
  end

  def unwatched(work)
    latest_status = @user.latest_statuses.find_by(work: work)

    return Episode.none if latest_status.blank?

    episode_ids = work.episodes.published.pluck(:id)
    work.episodes.where(id: (episode_ids - latest_status.watched_episode_ids))
  end
end
