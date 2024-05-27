# typed: false
# frozen_string_literal: true

module WorkRecordListSettable
  extend ActiveSupport::Concern

  def set_work_record_list(work)
    records = work
      .records_only_work
      .only_kept
      .eager_load(:work, :work_record, :episode_record, user: %i[gumroad_subscriber profile setting])
      .merge(WorkRecord.only_kept.order_by_rating(:desc).order(created_at: :desc))
    @my_records = @following_records = []

    if user_signed_in?
      records = records.where.not(user_id: current_user.mute_users.pluck(:muted_user_id))
      @my_records = records.merge(current_user.records.only_kept)
      @following_records = records.merge(current_user.followings)
      @all_records = records
        .where.not(user: [current_user, *current_user.followings])
        .merge(WorkRecord.with_body)
        .page(params[:page])
        .per(100)
    else
      @all_records = records
        .page(params[:page])
        .per(100)
    end
  end
end
