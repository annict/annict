class Marie::WorksController < Marie::ApplicationController
  permits :season_id, :title, :media, :official_site_url, :wikipedia_url,
          :released_at, :nicoch_started_at, :on_air, :twitter_username,
          :twitter_hashtag, :fetch_syobocal

  before_filter :set_work, only: [:show, :edit, :update]


  def index(page, q, season)
    @q = Work.search(q)
    @works = if q.present?
      @q.result.order(released_at: :desc).page(page)
    else
      Work.by_season(season).order(released_at: :desc).page(page)
    end
  end

  def on_air(page, q)
    @q = Work.search(q)
    @works = Work.on_air.order(released_at: :desc).page(page)

    render :index
  end

  def new
    @work = Work.new
  end

  def create(work)
    @work = Work.new(clean(work))

    if @work.save
      redirect_to marie_work_episodes_path(@work)
    else
      render 'new'
    end
  end

  def update(work)
    if @work.update_attributes(clean(work))
      redirect_to marie_work_path(@work)
    else
      render 'edit'
    end
  end


  private

  def set_work
    @work = Work.find(params[:id])
  end

  def clean(work)
    work[:released_at] = work[:released_at].strip.gsub(/年|月/, '-').delete('日')

    work
  end
end