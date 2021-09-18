# frozen_string_literal: true

module Fragment
  class WorkRecordsController < Fragment::ApplicationController
    include WorkRecordListSettable

    before_action :authenticate_user!

    def index
      work = Work.only_kept.find(params[:work_id])

      @form = WorkRecordForm.new(work: work, share_to_twitter: current_user.share_record_to_twitter?)

      set_work_record_list(work)
    end
  end
end
