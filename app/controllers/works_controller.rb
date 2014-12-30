class WorksController < ApplicationController
  include ApplicationHelper

  before_action :authenticate_user!, only: [:recommend]

  def index
    redirect_to on_air_works_path
  end

  def on_air(page: nil)
    @works = Work.on_air.order(watchers_count: :desc).page(page)

    @page_title = '現在放送中のアニメ一覧'
    @page_description = meta_description('現在放送中のアニメをチェック！')
    @page_keywords = meta_keywords('放送中', '今期')

    render :index
  end

  def popular(page: nil)
    @works = Work.order(watchers_count: :desc).page(page)

    @page_title = '人気アニメ一覧'
    @page_description = meta_description('Annictユーザに人気のアニメをチェック！')
    @page_keywords = meta_keywords('人気', '評判')

    render :index
  end

  def recommend(page: nil)
    work_ids = current_user.recommended_works(100).map(&:id)
    @works = current_user.works.unknown.where(id: work_ids)
      .order(watchers_count: :desc)
      .page(page)

    @page_title = "あなたにオススメのアニメ一覧"

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
    @episodes = @work.episodes.order(:sort_number)
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
