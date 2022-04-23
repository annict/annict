# frozen_string_literal: true

class RecordsController < ApplicationV6Controller
  include Pundit::Authorization
  include RecordListSettable

  before_action :authenticate_user!, only: %i[destroy]

  def index
    set_page_category PageCategory::RECORD_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile
    @dates = @user.records.only_kept.group_by_month(:watched_at).count.to_a.reverse.to_h

    set_user_record_list(@user)
  end

  def show
    set_page_category PageCategory::RECORD

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile
    @dates = @user.records.only_kept.group_by_month(:watched_at).count.to_a.reverse.to_h

    @record = @user.records.only_kept.find(params[:record_id])
    @work = @record.work
    @work_ids = [@work.id]
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
      work_record = @record.work_record
      work_record_list_path(work_record.work_id)
    end

    redirect_to path, notice: t("messages._common.deleted")
  end
end
