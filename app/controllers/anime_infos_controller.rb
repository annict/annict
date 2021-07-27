# frozen_string_literal: true

class AnimeInfosController < ApplicationV6Controller
  def show
    set_page_category PageCategory::ANIME_INFO

    @anime = Anime.only_kept.find(params[:anime_id])
    @programs = @anime.programs.eager_load(:channel).only_kept.in_vod.merge(Channel.order(:sort_number))
  end
end
