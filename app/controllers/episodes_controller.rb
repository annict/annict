class EpisodesController < ApplicationController
  before_filter :set_work,         only: [:index, :show]
  before_filter :set_episode,      only: [:show]
  before_filter :set_checkin_user, only: [:show]


  def show
    @checkins = @episode.checkins.order(created_at: :desc)

    if @checkin_user.present?
      @user_checkins = @checkins.where(user: @checkin_user)
    end

    if user_signed_in?
      @checkin = @episode.checkins.new
      @checkin.set_shared_sns(current_user)
    end
  end

  private

  def set_episode
    @episode = @work.episodes.find(params[:id])
  end

  def set_checkin_user
    if params[:username].present?
      @checkin_user = User.find_by(username: params[:username])
    end
  end
end
