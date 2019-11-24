# frozen_string_literal: true

class UserEpisodesQuery
  # @param user [User]
  # @param episodes [Episode::ActiveRecord_Relation]
  # @param status_kinds [Array<Symbol>]
  # @param watched [Boolean, nil]
  #
  # @return [Episode::ActiveRecord_Relation]
  def initialize(
    user,
    episodes,
    status_kinds: LatestStatus.kind.values.map(&:to_sym),
    watched: nil
  )
    @user = user
    @episodes = episodes
    @status_kinds = status_kinds
    @watched = watched
  end

  def call
    collection = user_episodes

    return collection if watched.nil?

    collection = filter_by_watched_episode_ids(collection) if watched
    collection = filter_by_unwatched_episode_ids(collection) unless watched

    collection
  end

  def slot(episode)
    channel_work = user.channel_works.find_by(work: episode.work)

    return if channel_work.blank?

    Slot.
      where(channel: channel_work.channel, episode: episode).
      without_deleted.
      order(started_at: :desc).
      first
  end

  def slot_data(latest_statuses)
    channel_works = user.channel_works.where(work_id: latest_statuses.pluck(:work_id))
    channel_ids = channel_works.pluck(:channel_id)
    episode_ids = latest_statuses.pluck(:next_episode_id)
    slots = Slot.
      includes(:channel, work: :work_image).
      where(channel_id: channel_ids, episode_id: episode_ids).
      without_deleted

    channel_works.map do |cw|
      slot = slots.
        select { |p| p.work_id == cw.work_id && p.channel_id == cw.channel_id }.
        sort_by(&:started_at).
        reverse.
        first

      slot
    end
  end

  private

  attr_reader :user, :episodes, :status_kinds, :watched

  def user_episodes
    return episodes.none if latest_statuses.blank?

    episodes.where(work_id: latest_statuses.pluck(:work_id))
  end

  def latest_statuses
    @latest_statuses ||= user.latest_statuses.with_kind(*status_kinds)
  end

  def filter_by_watched_episode_ids(collection)
    collection.where(id: watched_episode_ids)
  end

  def filter_by_unwatched_episode_ids(collection)
    collection.where(id: collection.pluck(:id) - watched_episode_ids)
  end

  def watched_episode_ids
    latest_statuses.pluck(:watched_episode_ids).flatten
  end
end
