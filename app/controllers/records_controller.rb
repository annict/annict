# frozen_string_literal: true

class RecordsController < ApplicationV6Controller
  include Pundit

  before_action :authenticate_user!, only: %i[destroy]

  def index
    set_page_category PageCategory::RECORD_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile
    @dates = @user.records.only_kept.group_by_month(:created_at).count.to_a.reverse.to_h

    @records = @user
      .records
      .preload(:anime_record, anime: :anime_image, episode_record: :episode)
      .order(created_at: :desc)
      .page(params[:page])
      .per(30)
      .without_count
    @records = @records.by_month(params[:month], year: params[:year]) if params[:month] && params[:year]
    @anime_ids = @records.pluck(:work_id)
  end

  def show
    set_page_category PageCategory::RECORD

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile
    @dates = @user.records.only_kept.group_by_month(:created_at).count.to_a.reverse.to_h

    @record = @user.records.only_kept.find(params[:record_id])
    @anime_ids = [@record.work_id]
  end

  def destroy
    @user = User.only_kept.find_by!(username: params[:username])
    @record = @user.records.only_kept.find(params[:record_id])

    authorize(@record, :destroy?)

    Destroyers::RecordDestroyer.new(record: @record).call

    path = if @record.episode_record?
      episode_record = @record.episode_record
      episode_path(episode_record.work_id, episode_record.episode_id)
    else
      work_record = @record.anime_record
      anime_record_list_path(work_record.work_id)
    end

    redirect_to path, notice: t("messages._common.deleted")
  end
end
