# frozen_string_literal: true

module TrackableEpisodeListSettable
  extend ActiveSupport::Concern

  def set_trackable_episode_list
    @library_entries = current_user
      .library_entries
      .with_not_deleted_anime
      .watching
      .eager_load(anime: :anime_image, program: :channel)
      .merge(Anime.where(no_episodes: false))
      .order(:position)
      .page(params[:page])
      .per(10)
      .without_count

    episodes = Episode.only_kept.where(work_id: @library_entries.pluck(:work_id))
    @trackable_episodes = Episode.partitioned_episodes(
      episode_condition: ["id IN (?)", episodes.pluck(:id) - @library_entries.pluck(:watched_episode_ids).flatten],
      limit: 1
    )
    @slots = Slot.only_kept.where(episode_id: @trackable_episodes.pluck(:id)).select(:program_id, :episode_id, :started_at)
  end
end
