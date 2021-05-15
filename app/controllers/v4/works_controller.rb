# frozen_string_literal: true

module V4
  class WorksController < ApplicationController
    include ApplicationHelper
    include AnimeSidebarDisplayable

    before_action :authenticate_user!, only: %i[switch]
    before_action :set_display_option, only: %i[popular newest season]

    def show
      set_page_category PageCategory::WORK

      anime = Work.only_kept.find(params[:anime_id])

      result = Rails.cache.fetch(work_detail_work_cache_key(anime), expires_in: 3.hours) {
        AnimePage::AnimeRepository.new(graphql_client: graphql_client).execute(anime_id: anime.id)
      }
      @anime_entity = result.anime_entity

      load_vod_channel_entities(anime: anime, anime_entity: @anime_entity)
    end

    def index
      redirect_to season_works_path(ENV["ANNICT_CURRENT_SEASON"])
    end

    def popular
      set_page_category PageCategory::WORK_LIST_POPULAR

      @works = Work
        .only_kept
        .preload(:work_image)
        .order(watchers_count: :desc, id: :desc)
        .page(params[:page])
        .per(display_works_count)

      render_list
    end

    def newest
      set_page_category PageCategory::WORK_LIST_NEWEST

      @works = Work
        .only_kept
        .preload(:work_image)
        .order(id: :desc)
        .page(params[:page])
        .per(display_works_count)

      render_list
    end

    def season
      set_page_category PageCategory::WORK_LIST_SEASON

      @works = Work
        .only_kept
        .by_season(params[:slug])
        .preload(:work_image)
        .order(watchers_count: :desc, id: :desc)
        .page(params[:page])
        .per(display_works_count)

      @seasons = Season.list(sort: :desc, include_all: true)
      @season = Season.find_by_slug(params[:slug])
      @prev_season = @season.sibling_season(:prev)
      @next_season = @season.sibling_season(:next)

      render_list
    end

    private

    def work_detail_work_cache_key(anime)
      [
        "anime-detail",
        "anime",
        anime.id,
        anime.updated_at.rfc3339
      ].freeze
    end

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
        @trailers_data = Work.trailers_data(@works)
        @casts_data = Work.casts_data(@works)
        @staffs_data = Work.staffs_data(@works, major: true)
        @programs_data = Work.programs_data(@works, only_vod: true)
        @channels = Channel.only_kept.with_vod
      end

      set_user_data_fetcher_params(
        work_ids: @works.pluck(:id),
        display_option: @display_option
      )
    end
  end
end
