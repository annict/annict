# frozen_string_literal: true

class RecordsController < ApplicationController
  permits :episode_id, :comment, :shared_twitter, :shared_facebook, :rating_state

  impressionist actions: %i(show)

  before_action :authenticate_user!, only: %i(edit update destroy)
  before_action :load_user, only: %i(show edit update destroy)
  before_action :load_record, only: %i(show edit update destroy)

  def show
    @work = @record.work
    @episode = @record.episode
    @comments = @record.comments.order(created_at: :desc)
    @comment = Comment.new
    @is_spoiler = user_signed_in? && current_user.hide_record?(@record)
    store_page_params(work: @work)
  end

  def edit
    authorize @record, :edit?
    @work = @record.work
  end

  def update(record)
    authorize @record, :update?

    @record.modify_comment = true
    @record.detect_locale!(:comment)

    if @record.update(record)
      @record.update_share_record_status
      @record.share_to_sns
      path = record_path(@user.username, @record)
      redirect_to path, notice: t("messages.records.updated")
    else
      @work = @record.work
      render :edit
    end
  end

  def destroy
    authorize @record, :destroy?

    @record.destroy

    path = work_episode_path(@record.work, @record.episode)
    redirect_to path, notice: t("messages.records.deleted")
  end

  def redirect(provider, url_hash)
    case provider
    when "tw"
      record = Record.published.find_by!(twitter_url_hash: url_hash)

      log_messages = [
        "Twitterからのアクセス ",
        "remote_host: #{request.remote_host}, ",
        "remote_ip: #{request.remote_ip}, ",
        "remote_user: #{request.remote_user}"
      ]
      logger.info(log_messages.join)

      bots = TwitterBot.pluck(:name)
      no_bots = bots.map do |bot|
        request.user_agent.present? && !request.user_agent.include?(bot)
      end
      record.increment!(:twitter_click_count) if no_bots.all?

      redirect_to_user_record(record, provider: "twitter")
    when "fb"
      record = Record.published.find_by!(facebook_url_hash: url_hash)
      record.increment!(:facebook_click_count)

      redirect_to_user_record(record, provider: "facebook")
    else
      redirect_to root_path
    end
  end

  private

  def load_user
    @user = User.find_by!(username: params[:username])
  end

  def load_record
    @record = @user.records.find(params[:id])
  end

  def redirect_to_user_record(record, provider:)
    username = record.user.username
    utm = {
      utm_source: provider,
      utm_medium: "record_share",
      utm_campaign: username
    }

    redirect_to record_path(username, record, utm)
  end
end
