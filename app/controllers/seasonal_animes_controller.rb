# frozen_string_literal: true

class SeasonalAnimesController < ApplicationV6Controller
  include AnimeListable

  def index
    set_page_category PageCategory::SEASONAL_ANIME_LIST

    @animes = Anime
      .only_kept
      .by_season(params[:season_slug])
      .preload(:anime_image)
      .order(watchers_count: :desc, id: :desc)
      .page(params[:page])
      .per(display_works_count)
      .without_count

    @seasons = Season.list(sort: :desc, include_all: true)
    @season = Season.find_by_slug(params[:season_slug])
    @prev_season = @season.sibling_season(:prev)
    @next_season = @season.sibling_season(:next)

    set_resource_data(@animes)
  end
end
