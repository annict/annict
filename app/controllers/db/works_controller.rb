class Db::WorksController < Db::ApplicationController
  permits :season_id, :sc_tid, :title, :media, :official_site_url, :wikipedia_url,
          :twitter_username, :twitter_hashtag, :released_at, :released_at_about,
          :fetch_syobocal

  before_action :authenticate_user!, only: [:new, :create, :edit, :update, :destroy]

  def index(page: nil)
    @works = Work.order(id: :desc).page(page)
  end

  def season(page: nil, slug: ENV["ANNICT_CURRENT_SEASON"])
    @works = Work.by_season(slug).order(id: :desc).page(page)
    render :index
  end

  def resourceless(page: nil, name: "episode")
    @works = case name
             when "episode" then Work.where(episodes_count: 0)
             when "item" then Work.itemless.includes(:item)
             end
    @works = @works.order(watchers_count: :desc).page(page)
    render :index
  end

  def search(q: nil, page: nil)
    @works = if q.present?
      @q.result.order(id: :desc)
    else
      Work.none
    end
    @works = @works.page(page)

    render :index
  end

  def new
    @work = Work.new
    authorize @work, :new?
  end

  def create(work)
    @work = Work.new(format_params(work))
    authorize @work, :create?

    if @work.save_and_create_db_activity(current_user, "works.create")
      redirect_to edit_db_work_path(@work), notice: "作品を登録しました"
    else
      render :new
    end
  end

  def edit(id)
    @work = Work.find(id)
    authorize @work, :edit?
  end

  def update(id, work)
    @work = Work.find(id)
    authorize @work, :update?

    @work.attributes = format_params(work)
    if @work.save_and_create_db_activity(current_user, "works.update")
      redirect_to edit_db_work_path(@work), notice: "作品を更新しました"
    else
      render :edit
    end
  end

  def hide(id)
    @work = Work.find(id)
    authorize @work, :hide?

    @work.hide!

    redirect_to db_works_path, notice: "作品を非公開にしました"
  end

  def destroy(id)
    @work = Work.find(id)
    authorize @work, :destroy?

    @work.destroy

    redirect_to db_works_path, notice: "作品を削除しました"
  end

  private

  def format_params(work_params)
    released_at = work_params[:released_at].strip.gsub(/年|月/, '-').delete('日')
    work_params[:released_at] = released_at
    work_params
  end
end
