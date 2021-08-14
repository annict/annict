# frozen_string_literal: true

module WorkRecordListSettable
  extend ActiveSupport::Concern

  def set_work_record_list(work)
    records = work
      .records
      .eager_load(:work, user: %i[gumroad_subscriber profile setting])
      .only_kept
      .only_work_record
      .order_by_rating(:desc)
    @my_records = @following_records = []

    if user_signed_in?
      @my_records = records.merge(current_user.records.only_kept)
      @following_records = records.merge(current_user.followings)
      @all_records = records
        .where.not(user: [current_user, *current_user.followings])
        .with_body
        .page(params[:page])
        .per(20)
    else
      @all_records = records
        .with_body
        .page(params[:page])
        .per(20)
    end
  end
end
