class WorksController < ApplicationController
  before_filter :set_work, only: [:show, :edit, :update]
  before_filter :authenticate_user!, only: [:recommend]


  def index(page)
    @works = Work.on_air.order(released_at: :desc).page(page)
  end

  def popular(page, filter)
    @works = ('on_air' == filter) ? Work.where(on_air: true) : Work
    @works = @works.order(watchers_count: :desc).page(page)
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
    @works = Work.by_season(name).order(released_at: :desc).page(page)
    render :index
  end

  def show
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


  private

  def set_work
    @work = Work.find(params[:id])
  end
end
