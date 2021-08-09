# frozen_string_literal: true

module TrackableEpisodeListSettable
  extend ActiveSupport::Concern

  def set_trackable_episode_list
    @library_entries = current_user
      .library_entries
      .with_not_deleted_anime
      .watching
      .preload(:next_episode, :next_slot)
      .eager_load(anime: :anime_image, program: :channel)
      .merge(Anime.where(no_episodes: false))
      .order(:position)
      .page(params[:page])
      .per(20)
      .without_count
  end
end
