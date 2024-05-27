# typed: false
# frozen_string_literal: true

class PopularWorksController < ApplicationV6Controller
  include WorkListable

  def index
    set_page_category PageCategory::POPULAR_WORK_LIST

    @works = Work
      .only_kept
      .preload(:work_image)
      .order(watchers_count: :desc, id: :desc)
      .page(params[:page])
      .per(display_works_count)

    set_resource_data(@works)
  end
end
