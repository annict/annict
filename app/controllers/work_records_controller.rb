# frozen_string_literal: true

class WorkRecordsController < ApplicationController
  permits :comment, :shared_twitter, :shared_facebook, :rating_state, model_name: "Record"

  before_action :authenticate_user!, only: %i(create)

  def index(work_id, page: nil)
    @work = Work.published.find(work_id)
    load_records(page)

    return unless user_signed_in?

    @record = @work.records.new

    store_page_params(work: @work)
  end

  def create(work_id, record)
    @work = Work.published.find(work_id)
    @record = @work.records.new(record)
    ga_client.page_category = params[:page_category]

    service = NewRecordService.new(current_user, @record)
    service.page_category = params[:page_category]
    service.ga_client = ga_client
    service.via = "web"

    begin
      service.save!
      flash[:notice] = t("messages.records.created")
      redirect_to work_records_path(@work)
    rescue
      load_records
      store_page_params(work: @work)
      render "/work_records/index"
    end
  end

  private

  def load_records(page = nil)
    @records = @work.
        records.
        published.
        reviews.
        includes(user: :profile)
    @records = localable_resources(@records)
    @records = @records.order(created_at: :desc).page(page)
    @is_spoiler = user_signed_in? && @records.present? && current_user.hide_record?(@records.first)
  end
end
