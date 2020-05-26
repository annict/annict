# frozen_string_literal: true

class WorkRecordsController < ApplicationController
  include V4::GraphqlRunnable

  before_action :authenticate_user!, only: %i(create edit update destroy)

  def index
    @work = Work.only_kept.find(params[:work_id])
    load_work_records

    return unless user_signed_in?

    @work_record = @work.work_records.new
    @work_record.setup_shared_sns(current_user)
  end

  def create
    @work = Work.only_kept.find(params[:work_id])

    _, err = CreateWorkRecordRepository.new(
      graphql_client: graphql_client(viewer: current_user)
    ).create(work: @work, params: work_record_params)

    if err
      load_work_records

      @work_record = @work.work_records.new(work_record_params)
      @work_record.errors.add(:mutation_error, err.message)
      @work_record.setup_shared_sns(current_user)

      return render "work_records/index"
    end

    flash[:notice] = t("messages._common.post")

    redirect_to work_records_path(@work)
  end

  def edit
    @record = current_user.records.find(params[:id])
    @work_record = current_user.
      work_records.
      only_kept.
      where(work_id: params[:work_id]).
      find(@record.work_record&.id)
    @work = @work_record.work
    authorize @work_record, :edit?
  end

  def update
    @record = current_user.records.find(params[:id])
    @work_record = current_user.
      work_records.
      only_kept.
      where(work_id: params[:work_id]).
      find(@record.work_record&.id)
    @work = @work_record.work
    authorize @work_record, :update?

    @work_record.attributes = work_record_params
    @work_record.detect_locale!(:body)
    @work_record.modified_at = Time.now

    begin
      ActiveRecord::Base.transaction do
        @work_record.save!
        current_user.update_share_record_setting(@work_record.share_to_twitter == "1")
        current_user.share_work_record_to_twitter(@work_record)
      end
      flash[:notice] = t("messages._common.updated")
      redirect_to record_path(@work_record.user.username, @work_record.record)
    rescue
      render :edit
    end
  end

  private

  def work_record_params
    params.require(:work_record).permit(
      :body, :rating_animation_state, :rating_music_state, :rating_story_state,
      :rating_character_state, :rating_overall_state, :share_to_twitter
    )
  end

  def load_work_records
    @work_records = UserWorkRecordsQuery.new.call(
      work_records: @work.work_records,
      user: current_user
    )
    @work_records = @work_records.with_body
    @work_records = localable_resources(@work_records)

    if user_signed_in?
      @my_work_records = UserWorkRecordsQuery.new.call(
        work_records: current_user.work_records.where(work: @work),
        user: current_user
      ).order(created_at: :desc)
      @work_records = @work_records.
        where.not(user: current_user).
        where.not(user_id: current_user.mute_users.pluck(:muted_user_id))
    end

    @work_records = @work_records.order(created_at: :desc).page(params[:page])
    @is_spoiler = user_signed_in? && @work_records.present? && current_user.hide_work_record_body?(@work_records.first.work)
  end
end
