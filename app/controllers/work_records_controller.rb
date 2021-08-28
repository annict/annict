# frozen_string_literal: true

class WorkRecordsController < ApplicationV6Controller
  include WorkRecordListSettable
  include WorkHeaderLoadable

  before_action :authenticate_user!, only: %i[edit update destroy]

  def index
    set_page_category PageCategory::WORK_RECORD_LIST

    set_work_header_resources

    @form = Forms::WorkRecordForm.new(work: @work, share_to_twitter: current_user&.share_record_to_twitter?)

    set_work_record_list(@work)
  end

  def edit
    set_page_category Rails.configuration.page_categories.record_edit

    @record = current_user.records.find(params[:id])
    @work_record = current_user
      .work_records
      .only_kept
      .where(work_id: params[:work_id])
      .find(@record.work_record&.id)
    @work = @work_record.work
    authorize @work_record, :edit?
  end

  def update
    @record = current_user.records.find(params[:id])
    @work_record = current_user
      .work_records
      .only_kept
      .where(work_id: params[:work_id])
      .find(@record.work_record&.id)
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
end
