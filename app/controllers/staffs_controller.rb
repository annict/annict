# frozen_string_literal: true

class StaffsController < ApplicationV6Controller
  include AnimeHeaderLoadable

  def index
    set_page_category PageCategory::STAFF_LIST

    set_anime_header_resources

    @staffs = @anime.staffs.preload(:resource).only_kept.order(:sort_number)
  end
end
