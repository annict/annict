class EpisodesController < ApplicationController
  before_filter :set_work,    only: [:index, :show]
  before_filter :set_episode, only: [:show]


  def show
    @checkins = @episode.checkins.order(created_at: :desc)

    if user_signed_in?
      @checkin = @episode.checkins.new
      @checkin.set_shared_sns(current_user)
    end
  end


  private

  def set_episode
    @episode = @work.episodes.find(params[:id])
  end
end
