class EpisodesController < ApplicationController
  before_filter :set_work,    only: [:index, :show]
  before_filter :set_episode, only: [:show]


  def show
    @checkins = @episode.checkins.order(created_at: :desc)
    @checkin = @episode.checkins.new if user_signed_in?
  end


  private

  def set_episode
    @episode = @work.episodes.find(params[:id])
  end
end