# frozen_string_literal: true

class ShareWorkRecordToTwitterJob < ApplicationJob
  queue_as :default

  def perform(user_id, work_record_id)
    user = User.find(user_id)
    work_record = user.work_records.only_kept.find(work_record_id)

    Deprecated::TwitterService.new(user).share!(work_record)
  end
end
