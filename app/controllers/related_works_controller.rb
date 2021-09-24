# frozen_string_literal: true

class RelatedWorksController < ApplicationV6Controller
  include WorkHeaderLoadable

  def index
    set_page_category PageCategory::RELATED_WORK_LIST

    set_work_header_resources

    @series_list = @work
      .series_list
      .eager_load(series_works: {work: :work_image})
      .only_kept
      .order("works.season_year ASC")
      .order("works.season_name ASC")
      .order(:created_at)

    @work_ids = [@work.id]
    @work_ids += @series_list.map do |series|
      series.series_works.pluck(:work_id)
    end.flatten
  end
end
