class Db::EpisodesController < Db::ApplicationController
  before_action :load_work, only: [:index, :destroy]

  def index
    @episodes = @work.episodes.order(:sort_number)
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
