# frozen_string_literal: true

class SeasonalAnimesController < ApplicationV6Controller
  before_action :set_display_option

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
    @anime_ids = @animes.pluck(:id)

    @seasons = Season.list(sort: :desc, include_all: true)
    @season = Season.find_by_slug(params[:season_slug])
    @prev_season = @season.sibling_season(:prev)
    @next_season = @season.sibling_season(:next)

    set_resource_data(@animes)
  end

  private

  def set_display_option
    @display_option = params[:display].in?(%w[grid grid_small]) ? params[:display] : "grid"
  end

  def display_works_count
    @display_option == "grid" ? 60 : 120
  end

  def set_resource_data(animes)
    if @display_option == "grid"
      @casts_data = Cast.only_kept.where(anime: animes).order(:sort_number).group_by(&:work_id)
      @staffs_data = Staff.only_kept.major.where(anime: animes).order(:sort_number).group_by(&:work_id)
    end
  end
end
