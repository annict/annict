# frozen_string_literal: true

class AnimesController < ApplicationV6Controller
  def show
    set_page_category PageCategory::ANIME

    @anime = Anime.only_kept.find(params[:anime_id])
    @trailers = @anime.trailers.only_kept.order(:sort_number).first(5)
    @episodes = @anime.episodes.only_kept.order(:sort_number).first(29)
    @programs = @anime.programs.eager_load(:channel).only_kept.in_vod.merge(Channel.order(:sort_number))
    @records = @anime
      .records
      .with_anime_record
      .only_kept
      .merge(AnimeRecord.with_body)
      .preload(:anime, :anime_record, :episode_record, user: %i[gumroad_subscriber profile])
      .order_by_rating(:desc)
      .first(11)
  end
end
