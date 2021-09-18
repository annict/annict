# frozen_string_literal: true

class WorkRecordsController < ApplicationV6Controller
  include WorkRecordListSettable
  include WorkHeaderLoadable

  def index
    set_page_category PageCategory::WORK_RECORD_LIST

    set_work_header_resources

    @form = WorkRecordForm.new(work: @work, share_to_twitter: current_user&.share_record_to_twitter?)

    set_work_record_list(@work)
  end
end
