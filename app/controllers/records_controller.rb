# frozen_string_literal: true

class RecordsController < ApplicationController
  permits :comment, :shared_twitter, :shared_facebook, :rating, model_name: "Checkin"

  before_action :authenticate_user!, only: %i(edit update destroy)
  before_action :load_user, only: %i(show edit update destroy)
  before_action :load_record, only: %i(show edit update destroy)

  def show
    @work = @record.work
    @episode = @record.episode
    @comments = @record.comments.order(created_at: :desc)
    @comment = Comment.new
  end

  def edit
    authorize @record, :edit?
    @work = @record.work
  end

  def update(checkin)
    authorize @record, :update?

    @record.modify_comment = true

    if @record.update_attributes(checkin)
      @record.update_share_checkin_status
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

  private

  def load_user
    @user = User.find_by(username: params[:username])
  end

  def load_record
    @record = @user.records.find(params[:id])
  end
end
