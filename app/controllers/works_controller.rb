# frozen_string_literal: true

class WorksController < ApplicationV6Controller
  include ApplicationHelper

  before_action :authenticate_user!, only: %i[switch]
  before_action :set_display_option, only: %i[popular newest season]

  def popular
    set_page_category PageCategory::WORK_LIST_POPULAR

    @works = Anime
      .only_kept
      .preload(:anime_image)
      .order(watchers_count: :desc, id: :desc)
      .page(params[:page])
      .per(display_works_count)

    render_list
  end

  def newest
    set_page_category PageCategory::WORK_LIST_NEWEST

    @works = Anime
      .only_kept
      .preload(:anime_image)
      .order(id: :desc)
      .page(params[:page])
      .per(display_works_count)

    render_list
  end

  def season
    set_page_category PageCategory::WORK_LIST_SEASON

    @works = Anime
      .only_kept
      .by_season(params[:slug])
      .preload(:anime_image)
      .order(watchers_count: :desc, id: :desc)
      .page(params[:page])
      .per(display_works_count)
      .without_count

    @seasons = Season.list(sort: :desc, include_all: true)
    @season = Season.find_by_slug(params[:slug])
    @prev_season = @season.sibling_season(:prev)
    @next_season = @season.sibling_season(:next)

    render_list
  end

  private

  def set_display_option
    display_options = Setting.display_option_work_list.values
    display = params[:display].in?(display_options) ? params[:display] : nil

    @display_option = display.presence || "list_detailed"
  end

  def display_works_count
    case @display_option
    when "list_detailed" then 15
    else
      120 # grid: 6 rows, grid_small: 10 rows
    end
  end

  def render_list
    if @display_option == "list_detailed"
      @trailers_data = Anime.trailers_data(@works)
      @casts_data = Anime.casts_data(@works)
      @staffs_data = Anime.staffs_data(@works, major: true)
      @programs_data = Anime.programs_data(@works, only_vod: true)
      @channels = Channel.only_kept.with_vod
    end
  end
end
