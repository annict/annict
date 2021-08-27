# frozen_string_literal: true

module WorkRecordListSettable
  extend ActiveSupport::Concern

  def set_work_record_list(work)
    records = work
      .records
      .eager_load(:work, user: %i[gumroad_subscriber profile setting])
      .only_kept
      .work_records
      .order_by_rating(:desc)
    @my_records = @following_records = @all_records = Record.none

    if user_signed_in?
      @my_records = records.merge(current_user.records.only_kept)
      @following_records = records.merge(current_user.followings)
      @all_records = records.where.not(user: [current_user, *current_user.followings])
    else
      @all_records = records
    end

    @all_records = @all_records
      .with_body
      .page(params[:page])
      .per(20)
  end
end
