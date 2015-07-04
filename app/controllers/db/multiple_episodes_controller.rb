class Db::MultipleEpisodesController < Db::ApplicationController
  permits :body

  before_action :load_work, only: [:new, :create]

  def new
    @multiple_episode = DB::MultipleEpisodesForm.new
    authorize @multiple_episode, :new?
  end

  def create(db_multiple_episodes_form)
    @multiple_episode = DB::MultipleEpisodesForm.new(db_multiple_episodes_form)
    authorize @multiple_episode, :create?

    @multiple_episode.work = @work

    if @multiple_episode.save
      flash[:notice] = "エピソードを登録しました"
      redirect_to db_work_episodes_path(@work)
    else
      render :new
    end
  end

  private

  def load_work
    @work = Work.find(params[:work_id])
  end
end
