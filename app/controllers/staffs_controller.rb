# frozen_string_literal: true

class StaffsController < ApplicationV6Controller
  include WorkHeaderLoadable

  def index
    set_page_category PageCategory::STAFF_LIST

    set_work_header_resources

    @staffs = @work.staffs.preload(:resource).only_kept.order(:sort_number)
  end
end
