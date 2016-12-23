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
  before_action :authenticate_user!, only: %i(create)
  before_action :load_work, only: %i(create show)
  before_action :load_episode, only: %i(create show)
  before_action :load_record, only: %i(show)

  def create(checkin)
    @record = @episode.checkins.new(checkin)
    service = NewRecordService.new(current_user, @record, ga_client)

    if service.save
      redirect_to work_episode_path(@work, @episode), notice: t("checkins.saved")
    else
      service = RecordsListService.new(@episode, current_user, nil)

      @record_user_ids = service.record_user_ids
      @user_records = service.user_records
      @current_user_records = service.current_user_records
      @records = service.records

      render "/episodes/show", layout: "v3/application"
    end
  end

  def show
    redirect_to record_path(@record.user.username, @record), status: 301
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

  def redirect_to_episode(checkin)
    work = checkin.episode.work
    username = checkin.user.username

    redirect_to work_episode_path(work, checkin.episode, username: username)
  end
end
