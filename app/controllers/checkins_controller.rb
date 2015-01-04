class CheckinsController < ApplicationController
  before_filter :authenticate_user!, only: [:destroy]
  before_filter :set_work,           only: [:show, :destroy]
  before_filter :set_episode,        only: [:show, :destroy]
  before_filter :set_checkin,        only: [:show, :destroy]
  before_filter :redirect_to_top,    only: [:destroy]

  def show
    @comments = @checkin.comments.order(created_at: :desc)
    @comment = Comment.new
  end

  def destroy
    @checkin.destroy
    redirect_to work_episode_path(@work, @episode), notice: t('checkins.deleted')
  end

  def redirect(provider, url_hash)
    if 'tw' == provider
      checkin = Checkin.find_by!(twitter_url_hash: url_hash)
      checkin.request_from_sns = true

      bots = TwitterBot.pluck(:name)
      no_bots = bots.map { |bot| request.user_agent.present? && !request.user_agent.include?(bot) }
      checkin.increment!(:twitter_click_count) if no_bots.all?

      redirect_to work_episode_checkin_path(checkin.episode.work, checkin.episode, checkin)
    elsif 'fb' == provider
      checkin = Checkin.find_by!(facebook_url_hash: url_hash)
      checkin.request_from_sns = true
      checkin.increment!(:facebook_click_count)

      redirect_to work_episode_checkin_path(checkin.episode.work, checkin.episode, checkin)
    else
      redirect_to root_path
    end
  end

  private

  def set_checkin
    @checkin = @episode.checkins.find(params[:id])
  end

  def redirect_to_top
    return redirect_to root_path if @checkin.user != current_user
  end
end
