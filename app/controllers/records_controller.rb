# frozen_string_literal: true

class RecordsController < ApplicationController
  permits :episode_id, :comment, :shared_twitter, :shared_facebook, :rating_state,
    model_name: "Checkin"

  before_action :authenticate_user!, only: %i(create edit update destroy switch)
  before_action :load_user, only: %i(create show edit update destroy)
  before_action :load_record, only: %i(show edit update destroy)

  def show
    @work = @record.work
    @episode = @record.episode
    @comments = @record.comments.order(created_at: :desc)
    @comment = Comment.new
    @is_spoiler = user_signed_in? && current_user.hide_checkin_comment?(@episode)
  end

  def create(checkin)
    @episode = Episode.published.find(checkin[:episode_id])
    @work = @episode.work
    @record = @episode.records.new(checkin)
    ga_client.page_category = params[:page_category]

    service = NewRecordService.new(current_user, @record)
    service.ga_client = ga_client

    begin
      service.save!
      flash[:notice] = t("messages.records.created")
      redirect_to work_episode_path(@work, @episode)
    rescue
      service = RecordsListService.new(current_user, @episode, params)

      @all_records = service.all_records
      @all_comment_records = service.all_comment_records
      @friend_comment_records = service.friend_comment_records
      @my_records = service.my_records
      @selected_comment_records = service.selected_comment_records

      data = {
        recordsSortTypes: Setting.records_sort_type.options,
        currentRecordsSortType: current_user&.setting&.records_sort_type.presence || "created_at_desc",
        pageObject: render_jb("works/_detail", user: current_user, work: @work)
      }
      gon.push(data)

      @is_spoiler = current_user.hide_checkin_comment?(@episode)

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

  def switch(episode_id, to)
    episode = Episode.find(episode_id)
    redirect = redirect_back fallback_location: work_episode_path(episode.work, episode)

    return redirect unless to.in?(Setting.display_option_record_list.values)

    current_user.setting.update_column(:display_option_record_list, to)
    redirect
  end

  private

  def load_user
    @user = User.find_by!(username: params[:username])
  end

  def load_record
    @record = @user.records.find(params[:id])
  end
end
