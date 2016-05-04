# frozen_string_literal: true
# == Schema Information
#
# Table name: works
#
#  id                :integer          not null, primary key
#  title             :string           not null
#  media             :integer          not null
#  official_site_url :string           default(""), not null
#  wikipedia_url     :string           default(""), not null
#  released_at       :date
#  created_at        :datetime
#  updated_at        :datetime
#  episodes_count    :integer          default(0), not null
#  season_id         :integer
#  twitter_username  :string
#  twitter_hashtag   :string
#  watchers_count    :integer          default(0), not null
#  sc_tid            :integer
#  released_at_about :string
#  aasm_state        :string           default("published"), not null
#  number_format_id  :integer
#  title_kana        :string           default(""), not null
#
# Indexes
#
#  index_works_on_aasm_state        (aasm_state)
#  index_works_on_episodes_count    (episodes_count)
#  index_works_on_media             (media)
#  index_works_on_number_format_id  (number_format_id)
#  index_works_on_released_at       (released_at)
#  index_works_on_sc_tid            (sc_tid) UNIQUE
#  index_works_on_watchers_count    (watchers_count)
#

class WorksController < ApplicationController
  include ApplicationHelper

  def index
    redirect_to season_works_path(ENV["ANNICT_CURRENT_SEASON"])
  end

  def popular(page: nil)
    @works = Work.published.order(watchers_count: :desc, id: :desc).page(page).per(15)

    render layout: "v1/application"
  end

  def season(slug, page: nil)
    @works = Work.
      published.
      by_season(slug).
      order(watchers_count: :desc, id: :desc).
      page(page).
      per(15)
    @season = Season.find_or_new_by_slug(slug)

    render layout: "v1/application"
  end

  def show
    @work = Work.published.find(params[:id])
    @status = current_user.latest_statuses.find_by(work: @work) if user_signed_in?

    render layout: "v1/application"
  end
end
