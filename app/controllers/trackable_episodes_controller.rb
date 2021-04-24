# frozen_string_literal: true

class TrackableEpisodesController < V4::ApplicationController
  include EpisodeRecordListLoadable

  before_action :authenticate_user!

  def index
    @library_entries = current_user.
      library_entries.
      with_not_deleted_work.
      wanna_watch_and_watching.
      eager_load(:work, program: :channel).
      merge(Anime.where(no_episodes: false)).
      order(:position).
      page(params[:page]).
      per(100)

    episodes = Episode.only_kept.where(work_id: @library_entries.pluck(:work_id))
    @untracked_episodes = Episode.partitioned_episodes(
      episode_condition: ["id IN (?)", episodes.pluck(:id) - @library_entries.pluck(:watched_episode_ids).flatten],
      limit: 3
    )
    @slots = Slot.only_kept.where(episode_id: @untracked_episodes.pluck(:id)).select(:program_id, :episode_id, :started_at)
  end

  def show
    @episode = Episode.only_kept.find(params[:episode_id])
    @anime = @episode.anime

    set_episode_record_list(@episode)
  end
end
