class Db::EpisodesController < Db::ApplicationController
  permits :number, :sort_number, :sc_count, :title, :prev_episode_id

  before_action :authenticate_user!
  before_action :load_work, only: [:index, :new, :create, :edit, :update, :destroy]

  def index
    @episodes = @work.episodes.order(:sort_number)
  end

  def edit(id)
    @episode = @work.episodes.find(id)
    authorize @episode, :edit?
  end

  def update(id, episode)
    @episode = @work.episodes.find(id)
    authorize @episode, :update?

    @episode.attributes = episode
    if @episode.save_and_create_db_activity(current_user, "episodes.update")
      redirect_to db_work_episodes_path(@work), notice: "エピソードを更新しました"
    else
      render :edit
    end
  end

  def destroy(id)
    @episode = @work.episodes.find(id)
    authorize @episode, :destroy?

    @episode.destroy

    redirect_to db_work_episodes_path(@work), notice: "エピソードを削除しました"
  end

  private

  def load_work
    @work = Work.find(params[:work_id])
  end
end
