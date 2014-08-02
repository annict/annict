class CheckinsController < ApplicationController
  permits :comment, :facebook_share, :spoil, :twitter_share

  before_filter :authenticate_user!, only: [:new, :create, :create_all, :edit, :update, :destroy]
  before_filter :set_work,           only: [:new, :create, :create_all, :show, :edit, :update, :destroy]
  before_filter :set_episode,        only: [:new, :create, :show, :edit, :update, :destroy]
  before_filter :set_checkin,        only: [:show, :edit, :update, :destroy]
  before_filter :redirect_to_top,    only: [:edit, :update, :destroy]


  def new
    @checkin = @episode.checkins.new
  end

  def create(checkin)
    @checkin = @episode.checkins.new(checkin)
    @checkin.user = current_user

    if @checkin.save
      redirect_to work_episode_path(@work, @episode), notice: t('checkins.saved')
    else
      render 'new'
    end
  end

  def create_all(episode_ids)
    if episode_ids =~ /\[([0-9]+,*)+\]/ # 括弧とカンマ、数字だけだったら
      episodes = eval(episode_ids).reverse.map { |id| Episode.find(id) }

      episodes.each do |episode|
        episode.checkins.create(user: current_user)
      end

      return redirect_to work_path(@work), notice: t('checkins.saved')
    end
  end

  def show
    @comments = @checkin.comments.order(created_at: :desc)
    @comment = Comment.new
  end

  def update(checkin)
    if @checkin.update_attributes(checkin)
      redirect_to work_episode_checkin_path(@work, @episode, @checkin), notice: t('checkins.updated')
    else
      render :edit
    end
  end

  def destroy
    @checkin.destroy
    redirect_to work_episode_path(@work, @episode), notice: t('checkins.deleted')
  end

  def redirect(provider, url_hash)
    if 'tw' == provider
      checkin = Checkin.find_by!(twitter_url_hash: url_hash)
      checkin.increment!(:twitter_click_count)

      redirect_to work_episode_checkin_path(checkin.episode.work, checkin.episode, checkin)
    elsif 'fb' == provider
      checkin = Checkin.find_by!(facebook_url_hash: url_hash)
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
