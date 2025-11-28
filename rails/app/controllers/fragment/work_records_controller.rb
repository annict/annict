# typed: false
# frozen_string_literal: true

module Fragment
  class WorkRecordsController < Fragment::ApplicationController
    include WorkRecordListSettable

    before_action :authenticate_user!, only: %i[index]

    def index
      work = Work.only_kept.find(params[:work_id])

      set_work_record_list(work)
    end
  end
end
