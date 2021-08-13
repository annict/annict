# frozen_string_literal: true

class VideosController < ApplicationV6Controller
  include WorkHeaderLoadable

  def index
    set_page_category PageCategory::VIDEO_LIST

    set_work_header_resources

    @videos = @work.trailers.only_kept.order(:sort_number)
  end
end
