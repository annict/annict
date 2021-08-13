# frozen_string_literal: true

class NewestWorksController < ApplicationV6Controller
  include WorkListable

  def index
    set_page_category PageCategory::NEWEST_WORK_LIST

    @works = Work
      .only_kept
      .preload(:work_image)
      .order(id: :desc)
      .page(params[:page])
      .per(display_works_count)

    set_resource_data(@works)
  end
end
