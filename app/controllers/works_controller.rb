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
end
