# frozen_string_literal: true

class EpisodesController < ApplicationV6Controller
  include EpisodeRecordListSettable
  include AnimeHeaderLoadable

  def index
    set_page_category PageCategory::EPISODE_LIST

    set_anime_header_resources
    raise ActionController::RoutingError, "Not Found" if @anime.no_episodes?

    @anime_ids = [@anime.id]
    @episodes = @anime.episodes.only_kept.order(:sort_number).page(params[:page]).per(100).without_count
  end

  def show
    set_page_category PageCategory::EPISODE

    @anime = Anime.only_kept.find(params[:anime_id])
    @anime_ids = [@anime.id]
    @programs = @anime.programs.eager_load(:channel).only_kept.in_vod.merge(Channel.order(:sort_number))

    @episode = @anime.episodes.only_kept.find(params[:episode_id])
    @form = Forms::EpisodeRecordForm.new(episode: @episode)

    set_episode_record_list(@episode)
  end
end
