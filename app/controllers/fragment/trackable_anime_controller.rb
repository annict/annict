# frozen_string_literal: true

module Fragment
  class TrackableAnimeController < Fragment::ApplicationController
    before_action :authenticate_user!

    def show
      @anime = Anime.only_kept.find(params[:anime_id])
      @library_entry = current_user.library_entries.find_by!(anime: @anime)
      @episodes = @anime
        .episodes
        .only_kept
        .where.not(id: @library_entry.watched_episode_ids)
        .order(:sort_number)
        .page(params[:page])
        .per(15)
      @programs = @anime
        .programs
        .only_kept
        .eager_load(:channel)
        .merge(current_user.channels.only_kept)
        .order(:started_at)
    end
  end
end
