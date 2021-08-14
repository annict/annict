# frozen_string_literal: true

class WorksController < ApplicationV6Controller
  include WorkHeaderLoadable

  def show
    set_page_category PageCategory::WORK

    set_work_header_resources

    @trailers = @work.trailers.only_kept.order(:sort_number).first(5)
    @episodes = @work.episodes.only_kept.order(:sort_number).first(29)
    @records = @work
      .records
      .preload(:work, user: %i[gumroad_subscriber profile])
      .only_kept
      .only_work_record
      .with_body
      .order_by_rating(:desc)
      .first(11)
  end
end
