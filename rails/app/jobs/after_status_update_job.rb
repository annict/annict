# typed: false
# frozen_string_literal: true

class AfterStatusUpdateJob < ApplicationJob
  queue_as :low

  def perform(user_id, work_id, prev_status_kind, new_status_kind)
    user = User.only_kept.find(user_id)
    work = Work.only_kept.find(work_id)

    ActiveRecord::Base.transaction do
      user.update_watched_works_count
      user.update_works_count!(prev_status_kind, new_status_kind)
      work.update_watchers_count!(prev_status_kind, new_status_kind)
    end
  end
end
