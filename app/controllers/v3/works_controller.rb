# frozen_string_literal: true

module V3
  class WorksController < V3::ApplicationController
    def show
      return
      @work = Work.published.find(params[:id])

      @casts = @work.
        casts.
        includes(:character, :person).
        published.
        order(:sort_number)

      @staffs = @work.
        staffs.
        includes(:resource).
        published.
        order(:sort_number)

      @channels = Channel.published.with_vod
      @series_list = @work.series_list.published.where("series_works_count > ?", 1)

      @work_records = UserWorkRecordsQuery.new.call(
        work_records: @work.work_records,
        user: current_user
      )
      @work_records = localable_resources(@work_records.with_body)
      @work_records = @work_records.order(created_at: :desc)

      @items = @work.items.published
      @items = localable_resources(@items)
      @items = @items.order(created_at: :desc).limit(10)

      return unless user_signed_in?

      store_page_params(work: @work)
    end
  end
end
