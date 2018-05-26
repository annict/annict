# frozen_string_literal: true
# == Schema Information
#
# Table name: works
#
#  id                :integer          not null, primary key
#  season_id         :integer
#  sc_tid            :integer
#  title             :string(510)      not null
#  media             :integer          not null
#  official_site_url :string(510)      default(""), not null
#  wikipedia_url     :string(510)      default(""), not null
#  episodes_count    :integer          default(0), not null
#  watchers_count    :integer          default(0), not null
#  released_at       :date
#  created_at        :datetime
#  updated_at        :datetime
#  twitter_username  :string(510)
#  twitter_hashtag   :string(510)
#  released_at_about :string
#  aasm_state        :string           default("published"), not null
#  number_format_id  :integer
#  title_kana        :string           default(""), not null
#
# Indexes
#
#  index_works_on_aasm_state        (aasm_state)
#  index_works_on_number_format_id  (number_format_id)
#  works_sc_tid_key                 (sc_tid) UNIQUE
#  works_season_id_idx              (season_id)
#

class WorksController < ApplicationController
  include ApplicationHelper

  before_action :set_cache_control_headers, only: %i(popular newest season)
  before_action :authenticate_user!, only: %i(switch)
  before_action :set_display_option, only: %i(popular newest season)

  def index
    redirect_to season_works_path(ENV["ANNICT_CURRENT_SEASON"])
  end

  def popular(page: nil)
    @works = Work.
      published.
      includes(:work_image).
      order(watchers_count: :desc, id: :desc).
      page(page).
      per(display_works_count)

    render_list
  end

  def newest(page: nil)
    @works = Work.
      published.
      includes(:work_image).
      order(id: :desc).
      page(page).
      per(display_works_count)

    render_list
  end

  def season(slug, page: nil)
    @works = Work.
      published.
      by_season(slug).
      includes(:work_image).
      order(watchers_count: :desc, id: :desc).
      page(page).
      per(display_works_count)

    @seasons = Season.list(sort: :desc, include_all: true)
    @season = Season.find_by_slug(slug)
    @prev_season = @season.sibling_season(:prev)
    @next_season = @season.sibling_season(:next)

    render_list
  end

  def show
    @work = Work.published.find(params[:id])

    @casts = @work.
      casts.
      includes(:character, :person).
      published.
      order(:sort_number)

    @staffs = @work.
      staffs.
      includes(:resource).
      published.
      order(:sort_number)

    @channels = Channel.published.with_vod
    @series_list = @work.series_list.published.where("series_works_count > ?", 1)

    @reviews = @work.
      reviews.
      published.
      with_body.
      includes(user: :profile)
    @reviews = localable_resources(@reviews)
    @reviews = @reviews.order(created_at: :desc)

    @items = @work.items.published
    @items = localable_resources(@items)
    @items = @items.order(created_at: :desc).limit(10)

    return unless user_signed_in?

    store_page_params(work: @work)
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
      @pvs_data = Work.pvs_data(@works)
      @casts_data = Work.casts_data(@works)
      @staffs_data = Work.staffs_data(@works, major: true)
      @program_details_data = Work.program_details_data(@works, only_vod: true)
      @channels = Channel.published.with_vod
    end

    store_page_params(works: @works, display_option: @display_option)

    set_surrogate_key_header(page_category, @works.map(&:record_key))
  end
end
