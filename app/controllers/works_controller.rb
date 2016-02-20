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

  def index
    redirect_to season_works_path(ENV["ANNICT_CURRENT_SEASON"])
  end

  def popular(page: nil)
    @works = Work.published.order(watchers_count: :desc).page(page).per(15)
  end

  def season(slug, page: nil)
    @works = Work.
      published.
      by_season(slug).
      order(watchers_count: :desc).
      page(page).
      per(15)
    @season = Season.find_or_new_by_slug(slug)
  end

  def show
    @work = Work.published.find(params[:id])
    @status = current_user.latest_statuses.find_by(work: @work) if user_signed_in?
  end

  def search(q: nil, page: nil)
    @q = Work.published.search(q)

    @works = if q.present?
      @q.result.order_latest.page(page)
    else
      Work.none
    end
  end
end
