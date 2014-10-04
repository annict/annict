class WorksController < ApplicationController
  before_action :authenticate_user!, only: [:recommend]

  def index
    redirect_to on_air_works_path
  end

  def on_air(page)
    @works = Work.on_air.order(watchers_count: :desc).page(page)
    render :index
  end

  def popular(page)
    @works = Work.order(watchers_count: :desc).page(page)
    render :index
  end

  def recommend(page)
    work_ids = current_user.recommended_works(100).map(&:id)
    @works = current_user.unknown_works.where(id: work_ids)
      .order(watchers_count: :desc)
      .page(page)
    render :index
  end

  def season(page, name)
    @works = Work.by_season(name).order(watchers_count: :desc).page(page)
    render :index
  end

  def show
    @work = Work.find(params[:id])
    @status = current_user.status(@work) if user_signed_in?
  end

  def search(q, page)
    @q = Work.search(q)

    @works = if q.present?
      @q.result.order(released_at: :desc).page(page)
    else
      Work.none
    end
  end
end
