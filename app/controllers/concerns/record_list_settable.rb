# frozen_string_literal: true

module RecordListSettable
  extend ActiveSupport::Concern

  def set_user_record_list(user)
    @records = user
      .records
      .preload(:episode, work: :work_image)
      .only_kept
      .order(created_at: :desc)
      .page(params[:page])
      .per(30)
    @records = @records.by_month(params[:month], year: params[:year]) if params[:month] && params[:year]
    @work_ids = @records.pluck(:work_id)
  end
end
