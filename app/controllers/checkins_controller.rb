# frozen_string_literal: true
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
#  work_id              :integer          not null
#  rating               :float
#  multiple_record_id   :integer
#  oauth_application_id :integer
#
# Indexes
#
#  checkins_episode_id_idx                 (episode_id)
#  checkins_facebook_url_hash_key          (facebook_url_hash) UNIQUE
#  checkins_twitter_url_hash_key           (twitter_url_hash) UNIQUE
#  checkins_user_id_idx                    (user_id)
#  index_checkins_on_multiple_record_id    (multiple_record_id)
#  index_checkins_on_oauth_application_id  (oauth_application_id)
#  index_checkins_on_work_id               (work_id)
#

class CheckinsController < ApplicationController
  permits :comment, :shared_twitter, :shared_facebook, :rating

  before_action :authenticate_user!, only: [:create, :create_all, :edit,
                                            :update, :destroy]
  before_action :set_work, only: [:create, :create_all, :show, :edit,
                                  :update, :destroy]
  before_action :set_episode, only: [:create, :show, :edit, :update, :destroy]
  before_action :load_record, only: [:show, :edit, :update, :destroy]
  before_action :redirect_to_top, only: [:edit, :update, :destroy]

  def create(checkin)
    record = @episode.checkins.new(checkin)
    service = NewRecordService.new(current_user, record, ga_client)

    if service.save
      redirect_to work_episode_path(@work, @episode), notice: t("checkins.saved")
    else
      service = RecordsListService.new(@episode, current_user, nil)

      @record_user_ids = service.record_user_ids
      @user_records = service.user_records
      @current_user_records = service.current_user_records
      @records = service.records

      @record = record_service.record

      render "/episodes/show", layout: "v1/application"
    end
  end

  def create_all(episode_ids)
    records = MultipleRecordsService.new(current_user)
    records.delay.save!(episode_ids)
    ga_client.events.create("multiple_records", "create")
    redirect_to work_path(@work), notice: t("checkins.saved")
  end

  def show
    @comments = @record.comments.order(created_at: :desc)
    @comment = Comment.new

    render layout: "v1/application"
  end

  def edit
    render layout: "v1/application"
  end

  def update(checkin)
    @record.modify_comment = true

    if @record.update_attributes(checkin)
      @record.update_share_checkin_status
      @record.share_to_sns
      redirect_to work_episode_checkin_path(@work, @episode, @record), notice: t('checkins.updated')
    else
      render :edit
    end
  end

  def destroy
    @record.destroy
    redirect_to work_episode_path(@work, @episode), notice: t("checkins.deleted")
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

  def load_record
    @record = @episode.checkins.find(params[:id])
  end

  def redirect_to_top
    return redirect_to root_path if @record.user != current_user
  end

  def redirect_to_episode(checkin)
    work = checkin.episode.work
    username = checkin.user.username

    redirect_to work_episode_path(work, checkin.episode, username: username)
  end
end
