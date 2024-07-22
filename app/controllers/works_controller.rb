# typed: false
# frozen_string_literal: true

class WorksController < ApplicationV6Controller
  include WorkHeaderLoadable

  def show
    set_page_category PageCategory::WORK

    set_work_header_resources

    @trailers = @work.trailers.only_kept.order(:sort_number).first(5)
    @episodes = @work.episodes.only_kept.order(:sort_number).first(29)
    records = @work
      .records
      .with_work_record
      .only_kept
      .merge(WorkRecord.with_body.order_by_rating(:desc).order(created_at: :desc))
      .preload(:work, :work_record, :episode_record, user: %i[gumroad_subscriber profile])
    if user_signed_in?
      records = records.where.not(user_id: current_user.mute_users.pluck(:muted_user_id))
    end
    @records = records.first(11)
  end
end
