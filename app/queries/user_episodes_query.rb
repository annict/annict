# frozen_string_literal: true

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

  def watched(work)
    latest_status = @user.latest_statuses.find_by(work: work)

    return Episode.none if latest_status.blank?

    work.episodes.where(id: latest_status.watched_episode_ids)
  end

  def program(episode)
    channel_work = @user.channel_works.find_by(work: episode.work)
    return if channel_work.blank?
    Program.
      where(channel: channel_work.channel, episode: episode).
      published.
      order(started_at: :desc).
      first
  end
end
