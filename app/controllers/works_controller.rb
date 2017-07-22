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

  before_action :authenticate_user!, only: %i(switch)
  before_action :load_i18n, only: %i(show popular newest season)

  def index
    redirect_to season_works_path(ENV["ANNICT_CURRENT_SEASON"])
  end

  def popular(page: nil)
    @works = Work.
      published.
      order(watchers_count: :desc, id: :desc).
      page(page).
      per(display_works_count)

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_list",
      user: current_user,
      works: @works
  end

  def newest(page: nil)
    @works = Work.
      published.
      order(id: :desc).
      page(page).
      per(display_works_count)

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_list",
      user: current_user,
      works: @works
  end

  def season(slug, page: nil)
    @works = Work.
      published.
      by_season(slug).
      includes(:work_image, :staffs, casts: %i(person character)).
      order(watchers_count: :desc, id: :desc).
      page(page).
      per(display_works_count)
    @seasons = Season.list(sort: :desc, include_all: true)
    @season = Season.find_by_slug(slug)
    @prev_season = @season.sibling_season(:prev)
    @next_season = @season.sibling_season(:next)

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_list",
      user: current_user,
      works: @works
  end

  def show
    @work = Work.published.find(params[:id])
    @episodes = @work.episodes.published.order(:sort_number)
    @casts = @work.
      casts.
      published.
      order(:sort_number)
    @staffs = @work.
      staffs.
      published.
      order(:sort_number)
    @series_list = @work.series_list.published.where("series_works_count > ?", 1)
    @reviews = @work.reviews.published.order(created_at: :desc)
    @items = @work.items.published.order(created_at: :desc).limit(10)

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_detail",
      user: current_user,
      work: @work
  end

  def switch(to)
    redirect = redirect_back fallback_location: works_path

    return redirect unless to.in?(Setting.display_option_work_list.values)

    current_user.setting.update_column(:display_option_work_list, to)
    redirect
  end

  private

  def load_i18n
    keys = {
      "messages._components.collect_button_modal.added": nil,
      "messages._components.collect_button_modal.view_collection": nil
    }

    load_i18n_into_gon keys
  end
end
