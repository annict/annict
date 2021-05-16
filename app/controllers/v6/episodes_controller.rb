# frozen_string_literal: true

module V6
  class EpisodesController < V6::ApplicationController
    include V6::EpisodeRecordListSettable

    def show
      set_page_category PageCategory::EPISODE

      @anime = Anime.only_kept.find(params[:anime_id])
      @episode = @anime.episodes.only_kept.find(params[:episode_id])
      @vod_channels = Channel.only_kept.joins(:programs).merge(@anime.programs.only_kept.in_vod).order(:sort_number)
      @form = EpisodeRecordForm.new(episode: @episode)

      set_episode_record_list(@episode)
    end
  end
end
