class WorksController < ApplicationController
  include ApplicationHelper

  def index
    redirect_to season_works_path(ENV["ANNICT_CURRENT_SEASON"])
  end

  def popular(page: nil)
    @works = Work.order(watchers_count: :desc).page(page)

    @page_title = '人気アニメ一覧'
    @page_description = meta_description('Annictユーザに人気のアニメをチェック！')
    @page_keywords = meta_keywords('人気', '評判')

    render :index
  end

  def season(name, page: nil)
    @works = Work.by_season(name).order(watchers_count: :desc).page(page)

    season = Season.find_by(slug: name)
    @page_title = "#{season.name}アニメ一覧"
    @page_description = meta_description("#{season.name}アニメをチェック！")
    @page_keywords = meta_keywords(season.name, '人気', '評判')

    render :index
  end

  def show
    @work = Work.find(params[:id])
    @status = current_user.statuses.kind_of(@work) if user_signed_in?
  end

  def search(q: nil, page: nil)
    @q = Work.search(q)

    @works = if q.present?
      @q.result.order(released_at: :desc).page(page)
    else
      Work.none
    end
  end
end
