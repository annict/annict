# frozen_string_literal: true

class WorkRecordsController < ApplicationV6Controller
  include WorkRecordListSettable
  include WorkHeaderLoadable

  before_action :authenticate_user!, only: %i[edit update destroy]

  def index
    set_page_category PageCategory::WORK_RECORD_LIST

    set_work_header_resources

    @form = Forms::WorkRecordForm.new(work: @work)

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
      @work_record.save!
      flash[:notice] = t("messages._common.updated")
      redirect_to record_path(@work_record.user.username, @work_record.record)
    rescue
      render :edit
    end
  end
end
