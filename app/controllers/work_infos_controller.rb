# frozen_string_literal: true

class WorkInfosController < ApplicationV6Controller
  include WorkHeaderLoadable

  def show
    set_page_category PageCategory::WORK_INFO

    set_work_header_resources
  end
end
