# typed: false
# frozen_string_literal: true

class CastsController < ApplicationV6Controller
  include WorkHeaderLoadable

  def index
    set_page_category PageCategory::CAST_LIST

    set_work_header_resources

    @casts = @work.casts.only_kept.order(:sort_number)
  end
end
