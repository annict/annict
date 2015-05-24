class Db::EpisodesController < Db::ApplicationController
  before_action :set_work, only: [:index, :destroy]
  before_action :set_episode, only: [:destroy]

  def index
    @episodes = @work.episodes.order(:sort_number)
  end

  def destroy
    @episode.destroy
    redirect_to db_work_episodes_path(@work), notice: "エピソードを削除しました"
  end

  private

  def set_episode
    @episode = @work.episodes.find(params[:id])
  end
end
