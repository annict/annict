class Api::Private::ApplicationController < ActionController::Base
  private

  def set_work
    @work = Work.find(params[:work_id])
  end

  def set_episode
    @episode = @work.episodes.find(params[:episode_id])
  end
end
