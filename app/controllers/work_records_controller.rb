# frozen_string_literal: true

class WorkRecordsController < ApplicationController
  permits :body, :rating_animation_state, :rating_music_state, :rating_story_state,
          :rating_character_state, :rating_overall_state

  before_action :authenticate_user!, only: %i(create edit update destroy)

  def index(work_id, page: nil)
    @work = Work.published.find(work_id)
    load_work_records

    return unless user_signed_in?

    @work_record = @work.work_records.new

    store_page_params(work: @work)
  end

  def create(work_id, work_record, page: nil)
    @work = Work.published.find(work_id)
    @work_record = @work.work_records.new(work_record)
    @work_record.user = current_user
    current_user.setting.attributes = setting_params
    ga_client.page_category = params[:page_category]

    service = NewWorkRecordService.new(current_user, @work_record, current_user.setting)
    service.ga_client = ga_client
    service.page_category = params[:page_category]
    service.via = "web"

    begin
      service.save!
      flash[:notice] = t("messages._common.post")
      redirect_to work_records_path(@work)
    rescue
      load_work_records

      store_page_params(work: @work)

      render "work_records/index"
    end
  end

  private

  def setting_params
    params.require(:setting).permit(:share_review_to_twitter, :share_review_to_facebook)
  end

  def load_work_records
    @work_records = @work.
      work_records.
      published.
      with_body.
      includes(user: :profile)
    @work_records = localable_resources(@work_records)
    @work_records = @work_records.order(created_at: :desc).page(params[:page])

    @is_spoiler = @user.present? && @work_records.present? && @user.hide_work_record_body?(@work_records.first.work)
  end
end
