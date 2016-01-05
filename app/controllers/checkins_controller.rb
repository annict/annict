# == Schema Information
#
# Table name: checkins
#
#  id                   :integer          not null, primary key
#  user_id              :integer          not null
#  episode_id           :integer          not null
#  comment              :text
#  modify_comment       :boolean          default(FALSE), not null
#  twitter_url_hash     :string(510)
#  facebook_url_hash    :string(510)
#  twitter_click_count  :integer          default(0), not null
#  facebook_click_count :integer          default(0), not null
#  comments_count       :integer          default(0), not null
#  likes_count          :integer          default(0), not null
#  created_at           :datetime
#  updated_at           :datetime
#  shared_twitter       :boolean          default(FALSE), not null
#  shared_facebook      :boolean          default(FALSE), not null
#  work_id              :integer
#
# Indexes
#
#  checkins_episode_id_idx         (episode_id)
#  checkins_facebook_url_hash_key  (facebook_url_hash) UNIQUE
#  checkins_twitter_url_hash_key   (twitter_url_hash) UNIQUE
#  checkins_user_id_idx            (user_id)
#  index_checkins_on_work_id       (work_id)
#

class CheckinsController < ApplicationController
  permits :comment, :shared_twitter, :shared_facebook

  before_action :authenticate_user!, only: [:create, :create_all, :edit,
                                            :update, :destroy]
  before_action :set_work, only: [:create, :create_all, :show, :edit,
                                  :update, :destroy]
  before_action :set_episode, only: [:create, :show, :edit, :update, :destroy]
  before_action :set_checkin, only: [:show, :edit, :update, :destroy]
  before_action :redirect_to_top, only: [:edit, :update, :destroy]

  def create(checkin)
    @checkin = @episode.checkins.new(checkin)
    @checkin.user = current_user
    @checkin.work = @work

    if @checkin.save
      @checkin.update_share_checkin_status
      @checkin.share_to_sns(self)
      keen_client.records.create(@checkin)
      redirect_to work_episode_path(@work, @episode), notice: t("checkins.saved")
    else
      render :new
    end
  end

  def create_all(episode_ids)
    records = MultipleRecordsService.new(current_user)
    records.delay.save!(episode_ids)
    redirect_to work_path(@work), notice: t("checkins.saved")
  end

  def show
    @comments = @checkin.comments.order(created_at: :desc)
    @comment = Comment.new
  end

  def update(checkin)
    if @checkin.update_attributes(checkin)
      @checkin.update_share_checkin_status
      @checkin.share_to_sns(self)
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

      logger.info("Twitterからのアクセス remote_host: #{request.remote_host}, remote_ip: #{request.remote_ip}, remote_user: #{request.remote_user}")

      bots = TwitterBot.pluck(:name)
      no_bots = bots.map { |bot| request.user_agent.present? && !request.user_agent.include?(bot) }
      checkin.increment!(:twitter_click_count) if no_bots.all?

      redirect_to_episode(checkin)
    elsif 'fb' == provider
      checkin = Checkin.find_by!(facebook_url_hash: url_hash)
      checkin.increment!(:facebook_click_count)

      redirect_to_episode(checkin)
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

  def redirect_to_episode(checkin)
    work = checkin.episode.work
    username = checkin.user.username

    redirect_to work_episode_path(work, checkin.episode, username: username)
  end
end
