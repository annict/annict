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

  def slot(episode)
    channel_work = @user.channel_works.find_by(work: episode.work)
    return if channel_work.blank?
    Slot.
      where(channel: channel_work.channel, episode: episode).
      published.
      order(started_at: :desc).
      first
  end

  def slot_data(latest_statuses)
    channel_works = @user.channel_works.where(work_id: latest_statuses.pluck(:work_id))
    channel_ids = channel_works.pluck(:channel_id)
    episode_ids = latest_statuses.pluck(:next_episode_id)
    slots = Slot.
      includes(:channel, work: :work_image).
      where(channel_id: channel_ids, episode_id: episode_ids).
      published

    channel_works.map do |cw|
      slot = slots.
        select { |p| p.work_id == cw.work_id && p.channel_id == cw.channel_id }.
        sort_by(&:started_at).
        reverse.
        first

      slot
    end
  end
end
