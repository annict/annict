# frozen_string_literal: true

class RecordsController < ApplicationController
  before_action :authenticate_user!, only: %i(destroy)

  def destroy
    @user = User.only_kept.find_by!(username: params[:username])
    @record = @user.records.only_kept.find(params[:id])

    authorize @record, :destroy?

    @record.destroy

    path = if @record.episode_record?
      episode_record = @record.episode_record
      episode_path(anime_id: episode_record.work_id, episode_id: episode_record.episode_id)
    else
      work_record = @record.work_record
      anime_record_list_path(anime_id: work_record.work_id)
    end

    redirect_to path, notice: t("messages._common.deleted")
  end
end
