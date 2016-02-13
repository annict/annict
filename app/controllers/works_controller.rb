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
    @works = Work.published.order(watchers_count: :desc).page(page)

    @page_title = '人気アニメ一覧'
    @page_description = meta_description('Annictユーザに人気のアニメをチェック！')
    @page_keywords = meta_keywords('人気', '評判')
  end

  def season(slug, page: nil)
    @works = Work.published.by_season(slug).order(watchers_count: :desc).page(page)
    @season = Season.find_or_new_by_slug(slug)

    yearly_season_ja = @season.decorate.yearly_season_ja
    @page_title = "#{yearly_season_ja}アニメ一覧"
    @page_description = meta_description("#{yearly_season_ja}アニメをチェック！")
    @page_keywords = meta_keywords(yearly_season_ja, "人気", "評判")
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
