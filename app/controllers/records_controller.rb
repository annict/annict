# frozen_string_literal: true

class RecordsController < ApplicationController
  permits :episode_id, :comment, :shared_twitter, :shared_facebook, :rating,
    model_name: "Checkin"

  before_action :authenticate_user!, only: %i(create edit update destroy)
  before_action :load_user, only: %i(create show edit update destroy)
  before_action :load_record, only: %i(show edit update destroy)

  def show
    @work = @record.work
    @episode = @record.episode
    @comments = @record.comments.order(created_at: :desc)
    @comment = Comment.new
  end

  def create(checkin)
    @episode = Episode.published.find(checkin[:episode_id])
    @work = @episode.work
    @record = @episode.records.new(checkin)
    keen_client.page_category = params[:page_category]
    service = NewRecordService.new(current_user, @record, keen_client)

    if service.save
      flash[:notice] = t("messages.records.created")
      redirect_to work_episode_path(@work, @episode)
    else
      service = RecordsListService.new(@episode, current_user, nil)

      @user_records = service.user_records
      @current_user_records = service.current_user_records
      @records = service.records

      render "/episodes/show"
    end
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
    @user = User.find_by!(username: params[:username])
  end

  def load_record
    @record = @user.records.find(params[:id])
  end
end
