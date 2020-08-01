# frozen_string_literal: true

class UserEpisodesQuery
  # @param user [User]
  # @param episodes [Episode::ActiveRecord_Relation]
  # @param watched [Boolean, nil]
  #
  # @return [Episode::ActiveRecord_Relation]
  def initialize(
    user,
    episodes,
    watched: nil
  )
    @user = user
    @episodes = episodes
    @watched = watched
  end

  def call
    collection = user_episodes

    return collection if watched.nil?

    collection = filter_by_watched_episode_ids(collection) if watched
    collection = filter_by_unwatched_episode_ids(collection) unless watched

    collection
  end

  private

  attr_reader :user, :episodes, :watched

  def user_episodes
    return episodes.none if library_entries.blank?

    episodes.where(anime_id: library_entries.pluck(:work_id))
  end

  def library_entries
    @library_entries ||= user.library_entries
  end

  def filter_by_watched_episode_ids(collection)
    collection.where(id: watched_episode_ids)
  end

  def filter_by_unwatched_episode_ids(collection)
    collection.where(id: collection.pluck(:id) - watched_episode_ids)
  end

  def watched_episode_ids
    library_entries.pluck(:watched_episode_ids).flatten
  end
end
