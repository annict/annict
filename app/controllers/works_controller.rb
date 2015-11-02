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

    render :index
  end

  def season(slug, page: nil)
    @works = Work.published.by_season(slug).order(watchers_count: :desc).page(page)

    season = Season.find_by_slug(slug)
    yearly_season_ja = season.decorate.yearly_season_ja
    @page_title = "#{yearly_season_ja}アニメ一覧"
    @page_description = meta_description("#{yearly_season_ja}アニメをチェック！")
    @page_keywords = meta_keywords(season.name, '人気', '評判')

    render :index
  end

  def show
    @work = Work.published.find(params[:id])
    @status = current_user.statuses.kind_of(@work) if user_signed_in?
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
