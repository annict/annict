# frozen_string_literal: true

class EpisodesController < V4::ApplicationController
  include EpisodeRecordListSettable

  def show
    set_page_category PageCategory::EPISODE

    @anime = Anime.only_kept.find(params[:anime_id])
    @episode = @anime.episodes.only_kept.find(params[:episode_id])
    @vod_channels = Channel.only_kept.eager_load(:programs).merge(@anime.programs.only_kept.in_vod).order(:sort_number)

    set_episode_record_list(@episode)
  end
end
