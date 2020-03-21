# frozen_string_literal: true

class CreateWorkRecordActivityJob < ApplicationJob
  queue_as :default

  def perform(user_id, work_record_id)
    user = User.find(user_id)
    work_record = user.work_records.only_kept.find(work_record_id)

    Activity.create! do |a|
      a.user = user
      a.recipient = work_record.work
      a.trackable = work_record
      a.action = "create_work_record"
      a.work = work_record.work
      a.work_record = work_record
    end
  end
end
