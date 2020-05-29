# frozen_string_literal: true

class RecordsController < ApplicationController
  before_action :authenticate_user!, only: %i(destroy)

  def index
    @user = User.only_kept.find_by!(username: params[:username])
    @records = UserEpisodeRecordsQuery.new.call(
      episode_records: @user.episode_records,
      user: @user
    ).order(created_at: :desc).page(params[:page])

    return unless user_signed_in?

    works = Work.only_kept.where(id: @records.pluck(:work_id))
  end

  def show
    @user = User.only_kept.find_by!(username: params[:username])
    @record = @user.records.only_kept.find(params[:id])

    if @record.episode_record?
      @episode_record = UserEpisodeRecordsQuery.new.call(
        episode_records: EpisodeRecord.where(id: @record.episode_record),
        user: current_user
      ).first
      raise ActiveRecord::RecordNotFound unless @episode_record

      @work = @episode_record.work
      @episode = @episode_record.episode
      @comments = @episode_record.comments.order(created_at: :desc)
      @comment = @episode_record.comments.new
      @is_spoiler = user_signed_in? && current_user.hide_episode_record_body?(@episode)
      render "episode_records/show"
    else
      @work_record = UserWorkRecordsQuery.new.call(
        work_records: WorkRecord.where(id: @record.work_record),
        user: current_user
      ).first
      raise ActiveRecord::RecordNotFound unless @work_record

      @work = @work_record.work
      @is_spoiler = user_signed_in? && current_user.hide_work_record_body?(@work)
      @work_records = UserWorkRecordsQuery.new.call(
        work_records: @user.work_records.where.not(id: @work_record.id),
        user: current_user
      )
      render "work_records/show"
    end
  end

  def destroy
    @user = User.only_kept.find_by!(username: params[:username])
    @record = @user.records.only_kept.find(params[:id])

    authorize @record, :destroy?

    @record.destroy

    path = if @record.episode_record?
      episode_record = @record.episode_record
      work_episode_path(episode_record.work, episode_record.episode)
    else
      work_record = @record.work_record
      work_records_path(work_record.work)
    end

    redirect_to path, notice: t("messages._common.deleted")
  end
end
